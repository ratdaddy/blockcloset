package cradle


type Client struct {
	pool *Pool
}

func New(pool *Pool) *Client {
	return &Client{pool: pool}
}

func (c *Client) Close() error {
	return c.pool.Close()
}
