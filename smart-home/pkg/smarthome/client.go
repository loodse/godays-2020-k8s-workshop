package smarthome

type Client struct {
	shutterClient *ShutterClient
	lightClient   *LightClient
}

func NewClient() *Client {
	c := &Client{
		shutterClient: newShutterClient(),
		lightClient:   newLightClient(),
	}
	return c
}

func (c *Client) Shutters() *ShutterClient {
	return c.shutterClient
}

func (c *Client) Lights() *LightClient {
	return c.lightClient
}

func (c *Client) Close() {
	c.shutterClient.close()
	c.lightClient.close()
}

type ValidationError string

func (e ValidationError) Error() string {
	return string(e)
}
