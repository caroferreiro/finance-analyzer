package pdf2csvcli

// BankType is a custom type to represent the bank enum.
type BankType string

var (
	Santander   BankType = "santander"
	VisaPrisma  BankType = "visa-prisma"
	MercadoPago BankType = "mercadopago"

	validBanks = []BankType{
		Santander,
		VisaPrisma,
		MercadoPago,
	}
)
