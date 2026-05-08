package part

type service struct {
	partRepository       PartRepository
	compatibilityChecker CompatibilityChecker
}

func NewService(partRepository PartRepository, compatibilityChecker CompatibilityChecker) *service {
	return &service{
		partRepository:       partRepository,
		compatibilityChecker: compatibilityChecker,
	}
}
