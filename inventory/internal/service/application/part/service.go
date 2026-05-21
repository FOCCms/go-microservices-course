package part

type service struct {
	partRepository       PartRepository
	compatibilityChecker CompatibilityChecker
	txManager            TxManager
}

func NewService(partRepository PartRepository, compatibilityChecker CompatibilityChecker, txManager TxManager) *service {
	return &service{
		partRepository:       partRepository,
		compatibilityChecker: compatibilityChecker,
		txManager:            txManager,
	}
}
