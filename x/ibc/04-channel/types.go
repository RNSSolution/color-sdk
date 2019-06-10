package channel

type State = byte

const (
	Init State = iota
	OpenTry
	Open
	CloseTry
	Closed
)

type Packet struct {
	Sequence      uint64
	TimeoutHeight uint64

	SourceConnection string
	SourceChannel    string
	DestConnection   string
	DestChannel      string

	Data []byte
}

type Channel struct {
	Module             string
	Counterparty       string
	CounterpartyModule string
}
