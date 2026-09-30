package main

// Registry — реестр «ключ → значение» (адреса сервисов, фичефлаги) с
// подпиской на изменения. Get зовут очень часто, Set — редко.
type Registry struct {
	// ваши поля
}

func NewRegistry() *Registry {
	// ваш код
	return &Registry{}
}

// Get возвращает текущее значение ключа.
func (r *Registry) Get(key string) (string, bool) {
	// ваш код
	return "", false
}

// Set записывает значение и уведомляет подписчиков ключа. Set никогда не
// блокируется из-за медленного подписчика: если подписчик не успел
// прочитать прошлое уведомление, оно заменяется новым — подписчик
// может пропустить промежуточные значения, но последнее получит всегда.
func (r *Registry) Set(key, value string) {
	// ваш код
}

// Watch подписывается на изменения ключа. Если значение уже есть, оно
// сразу лежит в канале. cancel отписывает и закрывает канал; повторный
// вызов cancel ничего не делает, и его безопасно звать одновременно с Set.
func (r *Registry) Watch(key string) (updates <-chan string, cancel func()) {
	// ваш код
	return make(chan string), func() {}
}

// Watchers — сколько сейчас активных подписок на ключ.
func (r *Registry) Watchers(key string) int {
	// ваш код
	return 0
}
