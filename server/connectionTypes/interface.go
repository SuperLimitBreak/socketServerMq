package connectionTypes

type Connection interface {
	GetIngressChan() chan []byte
	GetEgressChan() chan []byte
	SetEgressChan(chan []byte)
}
