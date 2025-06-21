package atlas

type Atlas struct {
    ID          uint
    Name        string
    Description string
    Category    string // e.g., "Vegetal", "Animal", "Fungal", etc.
    TissueRecords []uint // IDs of associated TissueRecords
}