package microservice

func MicroServiceHealthyHandler(ctx IContext) error {
	ctx.Response(200, "OK")
	return nil
}
