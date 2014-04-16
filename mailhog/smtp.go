package mailhog

type Message struct {
	From string
	To string
	Data []byte
	Helo string
}
