package main

var (
	chkErrPay    = errors.New("карта отклонена")
	chkErrRefund = errors.New("возврат не прошёл")
)

type chkLog struct{ ops []string }

func (l *chkLog) step(name string, doErr, compErr error) Step {
	return Step{
		Name: name,
		Do:   func(context.Context) error { l.ops = append(l.ops, "do:"+name); return doErr },
		Compensate: func(ctx context.Context) error {
			if ctx.Err() != nil {
				l.ops = append(l.ops, "undo-canceled:"+name)
			} else {
				l.ops = append(l.ops, "undo:"+name)
			}
			return compErr
		},
	}
}

func TestSagaOK(t *testing.T) {
	var l chkLog
	err := Run(context.Background(), []Step{l.step("a", nil, nil), l.step("b", nil, nil)})
	if err != nil || !slices.Equal(l.ops, []string{"do:a", "do:b"}) {
		t.Fatalf("всё успешно: err=%#v ops=%v", err, l.ops)
	}
}

func TestSagaCompensates(t *testing.T) {
	var l chkLog
	noUndo := Step{Name: "log", Do: func(context.Context) error { l.ops = append(l.ops, "do:log"); return nil }}
	err := Run(context.Background(), []Step{
		l.step("reserve", nil, nil), noUndo, l.step("hold", nil, nil), l.step("pay", fmt.Errorf("банк: %w", chkErrPay), nil), l.step("ship", nil, nil),
	})
	want := []string{"do:reserve", "do:log", "do:hold", "do:pay", "undo:hold", "undo:reserve"}
	if !slices.Equal(l.ops, want) {
		t.Fatalf("операции %v\nожидали %v (откат выполненных в обратном порядке, упавший не откатывается)", l.ops, want)
	}
	var se *SagaError
	if !errors.As(err, &se) || se.Step != "pay" || !errors.Is(err, chkErrPay) {
		t.Fatalf("ожидали *SagaError на шаге pay с причиной, получили %v", err)
	}
	if err.Error() != "сага: шаг pay: банк: карта отклонена" {
		t.Fatalf("текст %q", err.Error())
	}
}

func TestSagaCompensationErrors(t *testing.T) {
	var l chkLog
	err := Run(context.Background(), []Step{
		l.step("a", nil, errors.New("не откатили a")), l.step("b", nil, chkErrRefund), l.step("c", chkErrPay, nil),
	})
	if !slices.Equal(l.ops, []string{"do:a", "do:b", "do:c", "undo:b", "undo:a"}) {
		t.Fatalf("ошибка компенсации не должна останавливать остальные: %v", l.ops)
	}
	var se *SagaError
	if !errors.As(err, &se) || len(se.Compensation) != 2 {
		t.Fatalf("ожидали 2 ошибки компенсаций, получили %v", err)
	}
	if se.Compensation[0].Error() != "компенсация b: возврат не прошёл" {
		t.Fatalf("первая ошибка компенсации %q, ожидали «компенсация b: возврат не прошёл»", se.Compensation[0])
	}
	if !errors.Is(err, chkErrRefund) || !errors.Is(err, chkErrPay) {
		t.Fatal("errors.Is должен находить и ошибку шага, и ошибки компенсаций")
	}
	if !strings.HasSuffix(err.Error(), "; компенсаций не удалось: 2") {
		t.Fatalf("текст %q должен сообщать о 2 неудачных компенсациях", err.Error())
	}
}

func TestSagaCanceled(t *testing.T) {
	var l chkLog
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cancelling := Step{Name: "b", Do: func(context.Context) error { l.ops = append(l.ops, "do:b"); cancel(); return nil },
		Compensate: func(ctx context.Context) error {
			if ctx.Err() != nil {
				l.ops = append(l.ops, "undo-canceled:b")
			} else {
				l.ops = append(l.ops, "undo:b")
			}
			return nil
		}}
	err := Run(ctx, []Step{l.step("a", nil, nil), cancelling, l.step("c", nil, nil)})
	want := []string{"do:a", "do:b", "undo:b", "undo:a"}
	if !slices.Equal(l.ops, want) {
		t.Fatalf("операции %v\nожидали %v (шаг c не выполняется, откат получает неотменённый контекст)", l.ops, want)
	}
	var se *SagaError
	if !errors.As(err, &se) || se.Step != "c" || !errors.Is(err, context.Canceled) {
		t.Fatalf("ожидали *SagaError на шаге c с context.Canceled, получили %v", err)
	}
}
