package adatft

// Es ist vorstellbar, den TFT-Display ueber andere Schnittstellen als SPI
// anzusteuern und andere Chips fuer die Ansteuerung als den ILI9341 zu
// verwenden. Dieses Interface beschreibt alle Methoden, welche
// von einer Display-Ansteuerung implementiert werden muessen. Da bisher
// nur der ILI9341 via SPI angesteuert wurde, ist es sehr wahrscheinlich,
// dass dieses Interface noch stark angepasst werden muss.
type DispInterface interface {
	// Damit die Initialisierung so flexibel wie moeglich bleibt, wird der
	// Init-Methode ein Slice von beliebigen Parametern uebergeben. Wie die
	// Werte interpretiert werden, ist Interface-spezifisch.
	Init(initParams []any)

	// Schliesst die Verbindung zum ILI-Chip und gibt alle Ressourcen in
	// Zusammenhang mit dieser Verbindung frei.
	Close()

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

// Wie für den Display, so gibt es auch für den Touchscreen-Controller
// verschiedene Ausführungen. Dieses Interface beschreibt alle Methoden,
// welche von einer Touchscreen-Anbindung implementiert werden müssen.
type TouchInterface interface {
	// Damit die Initialisierung so flexibel wie moeglich bleibt, wird der
	// Init-Methode ein Slice von beliebigen Parametern uebergeben. Wie die
	// Werte interpretiert werden, ist Interface-spezifisch.
	Init(initParams []any)

	// Schliesst die Verbindung zum Touchscreen-Controller und gibt alle
	// Ressourcen im Zusammenhang mit dieser Verbindung frei.
	Close()

	// Mit den folgenden vier Methoden können die Register des Controller
	// ausgelesen oder beschrieben werden. Es stehen Methoden für 8-Bit oder
	// 16-Bit Register zur Verfügung.
	ReadReg8(addr uint8) uint8
	WriteReg8(addr uint8, value uint8)
	ReadReg16(addr uint8) uint16
	WriteReg16(addr uint8, value uint16)

	// Mit ReadData kann die aktuelle Position auf dem Touchscreen ermittelt
	// werden. Diese Methode sollte nur dann aufgerufen werden, wenn auch
	// Positionsdaten vorhanden sind.
	ReadData() (x, y uint16, z uint8)

	// Damit wird die Funktion cbFunc als Handler für alle Interrupts im
	// Zusammenhang mit dem Touchscreen hinterlegt. Der Funktion wird beim
	// Aufruf der Paramter cbData uebergeben.
	SetCallback(cbFunc func(any), cbData any)
}
