package main

func chkCSV(t *testing.T, in string, want [][]string) {
	t.Helper()
	got, err := ReadCSV(strings.NewReader(in))
	if err != nil {
		t.Fatalf("ReadCSV(%q): неожиданная ошибка %v", in, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ReadCSV(%q) =\n%q\nожидали\n%q", in, got, want)
	}
}

func TestCSVPlain(t *testing.T) {
	chkCSV(t, "id,name,city\n1,Аня,Москва\r\n2,Bob,\n\n3,,Kazan", [][]string{
		{"id", "name", "city"}, {"1", "Аня", "Москва"}, {"2", "Bob", ""}, {"3", "", "Kazan"},
	})
	chkCSV(t, "", nil)
	chkCSV(t, "\n\r\n\n", nil)
}

func TestCSVQuoted(t *testing.T) {
	in := "name,comment\n" +
		"\"Иванов, Иван\",\"сказал \"\"привет\"\"\"\n" +
		"\"\",\"две\r\nстроки\"\r\n" +
		"x,\"a,b\"\n"
	chkCSV(t, in, [][]string{
		{"name", "comment"},
		{"Иванов, Иван", `сказал "привет"`},
		{"", "две\nстроки"},
		{"x", "a,b"},
	})
}

func chkCSVErr(t *testing.T, in string, target error, line int) {
	t.Helper()
	got, err := ReadCSV(strings.NewReader(in))
	var ce *CSVError
	if got != nil || !errors.As(err, &ce) || !errors.Is(err, target) || ce.Line != line {
		t.Fatalf("ReadCSV(%q) = %q, %v; ожидали nil и *CSVError{Line: %d} с %v", in, got, err, line, target)
	}
}

func TestCSVQuoteErrors(t *testing.T) {
	chkCSVErr(t, "a,b\nc,d\"e\n", ErrQuote, 2)
	chkCSVErr(t, "a,b\n\"c\"x,d\n", ErrQuote, 2)
	chkCSVErr(t, "a,b\n\"многострочное\nполе\",1\n\"не закрыта,2\nи дальше\n", ErrQuote, 4)
}

func TestCSVFieldCount(t *testing.T) {
	chkCSVErr(t, "a,b\n1,2\n\n\"x\ny\",2,3\n", ErrFieldCount, 4)
	chkCSV(t, "a,b,\n1,2,\n", [][]string{{"a", "b", ""}, {"1", "2", ""}})
}

type chkSlowReader struct {
	data string
	err  error
}

// Отдаёт по одному байту — руны и "" разрываются между вызовами Read.
func (r *chkSlowReader) Read(p []byte) (int, error) {
	if r.data == "" {
		return 0, r.err
	}
	p[0] = r.data[0]
	r.data = r.data[1:]
	return 1, nil
}

func TestCSVReaderBehaviour(t *testing.T) {
	got, err := ReadCSV(&chkSlowReader{data: "\"ё\"\"ж\",я\n", err: io.EOF})
	if err != nil || !reflect.DeepEqual(got, [][]string{{`ё"ж`, "я"}}) {
		t.Fatalf("побайтовое чтение: %q, %v", got, err)
	}
	disk := errors.New("ошибка диска")
	got, err = ReadCSV(&chkSlowReader{data: "a,b\n1,2", err: disk})
	if got != nil || !errors.Is(err, disk) {
		t.Fatalf("ошибка чтения потерялась: %q, %v", got, err)
	}
}
