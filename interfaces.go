package adatft

// Es ist möglich, verschiedene Libraries für die SPI-Anbindung des
// ILI-Chips zu verwenden. Dieses Interface beschreibt alle Funktionen, welche
// von einer SPI-Anbindung implementiert werden müssen.
type DispInterface interface {
	// Schliesst die Verbindung zum ILI-Chip und gibt alle Ressourcen in
	// Zusammenhang mit dieser Verbindung frei.
	Close()

	Init(initParams []any)

	// Sendet einen Befehl (Command) zum Chip. Das ist in der Regel ein
	// 8 Bit Wert.
	Cmd(cmd uint8)

	// Sendet 8 Bit als Daten zum Chip. In den meisten Fällen ist dies ein
	// Argument eines Befehls, der vorgängig via Cmd gesendet wird.
	Data8(val uint8)

	// Analog Data8, jedoch mit 32 Bit Daten.
	Data32(val uint32)

	// Der gesamte Slice buf wird gesendet.
	DataArray(buf []byte)
}
