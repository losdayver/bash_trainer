package handlers

/*
	RECEIVED TYPES
*/

// JSON is received when user tries to gets commands for palette
type GetCommandPacket struct {
	Username  string
	UserToken string
}

type SaveCommandPacket struct {
	Command   string
	Username  string
	UserToken string
}

/*
	SENT TYPES
*/

// JSON is sent to user who requested commands for their palette
type CommandPalettePacket struct {
	Commands []string
}
