package miner

import "testing"

func TestCompileDifficultyBits(t *testing.T) {
	m := InitMiner(nil, nil, 1)

	m.CompileDifficultyBits(20)
	if m.difficultyBits != 20 {
		t.Fatalf("expected difficulty 20, got %d", m.difficultyBits)
	}

	m.CompileDifficultyBits(-5)
	if m.difficultyBits != 0 {
		t.Fatalf("expected difficulty clamped to 0, got %d", m.difficultyBits)
	}
}
