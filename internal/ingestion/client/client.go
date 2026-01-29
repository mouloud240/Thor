package client
type Client struct{
	Req string
	ResChan chan<- string
}
func NewClient(req string,resChan chan<- string)*Client{
     return &Client{Req: req,ResChan: resChan}
}

