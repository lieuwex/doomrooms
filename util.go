package main

func PlayerIndex(players []*Player, player *Player) int {
	for i, p := range players {
		if p == player {
			return i
		}
	}
	return -1
}
