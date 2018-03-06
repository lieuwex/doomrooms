package main

import "os"

func PlayerIndex(players []*Player, player *Player) int {
	for i, p := range players {
		if p == player {
			return i
		}
	}
	return -1
}

func GameServerIndex(servers []*GameServer, gs *GameServer) int {
	for i, val := range servers {
		if val == gs {
			return i
		}
	}
	return -1
}

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	// other error
	return false, err
}
