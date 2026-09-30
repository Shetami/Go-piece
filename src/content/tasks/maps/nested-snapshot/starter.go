package main

// Registry — потокобезопасный реестр настроек: раздел → ключ → значение.
//
//   - Set пишет значение, создавая раздел при необходимости.
//   - Get читает; ok=false, если нет раздела или ключа.
//   - Delete удаляет ключ; раздел, в котором не осталось ключей, исчезает.
//   - Snapshot возвращает независимую копию всего реестра: правки снимка
//     (в том числе внутри разделов) не видны реестру, а правки реестра —
//     снимку. У пустого реестра снимок — пустая, но не nil мапа.
type Registry struct {
	// ваши поля
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Set(section, key, value string) {
	// ваш код
}

func (r *Registry) Get(section, key string) (string, bool) {
	// ваш код
	return "", false
}

func (r *Registry) Delete(section, key string) {
	// ваш код
}

func (r *Registry) Snapshot() map[string]map[string]string {
	// ваш код
	return nil
}
