package message

type GetTxs struct{}

type GetCurrentHeader struct{}

type GetTransaction struct{
	Key []byte
}
