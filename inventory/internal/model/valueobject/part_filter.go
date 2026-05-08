package valueobject

type PartFilter struct {
	// UUIDs — если не пустой, возвращаются только эти детали (приоритет)
	UUIDs []string
	// PartType — фильтр по типу (игнорируется если UUIDs заполнен)
	PartType PartType
}
