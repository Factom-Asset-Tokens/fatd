package fat0

type Transaction struct {
	Entry
}

func (t *Transaction) Unmarshal() error {
	return t.Entry.Unmarshal(t)
}
