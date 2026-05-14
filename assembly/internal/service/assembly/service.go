package assembly

type service struct {
	assemblyProducerService AssemblyProducerService
}

func NewService(assemblyProducerService AssemblyProducerService) *service {
	return &service{
		assemblyProducerService: assemblyProducerService,
	}
}
