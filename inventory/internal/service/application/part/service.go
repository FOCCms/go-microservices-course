package part

type service struct {
	partRepository PartRepository
}

func NewService(partRepository PartRepository) *service {
	return &service{
		partRepository: partRepository,
	}
}
