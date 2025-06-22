package utils

type CandleWindow struct {
	size int
	data []Kline
}

// NewCandleWindow создаёт пустой буфер для size свечей.
func NewCandleWindow(size int) *CandleWindow {
	return &CandleWindow{
		size: size,
		data: make([]Kline, 0, size),
	}
}

// Load инициализирует окно начальными данными (REST).
func (w *CandleWindow) Load(initial []Kline) {
	if len(initial) > w.size {
		// берем только последние size элементов
		w.data = append([]Kline(nil), initial[len(initial)-w.size:]...)
	} else {
		w.data = append([]Kline(nil), initial...)
	}
}

// Add добавляет новую свечу и, при переполнении, удаляет самую старую.
func (w *CandleWindow) Add(c Kline) {
	if len(w.data) < w.size {
		w.data = append(w.data, c)
		return
	}
	// сдвигаем окно: отбрасываем первый элемент
	copy(w.data[0:], w.data[1:])
	w.data[w.size-1] = c
}

// Data возвращает срез из последних свечей (не копируя, для скорости)
func (w *CandleWindow) Data() []Kline {
	return w.data
}
