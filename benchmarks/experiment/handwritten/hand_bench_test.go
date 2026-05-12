package handwritten

import "testing"

// ============================================================
// Эксперимент А — Варьирование размера payload
// ============================================================

func BenchmarkHandBuild_64B_0ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BuildPacket(0, 64)
	}
}
func BenchmarkHandParse_64B_0ext(b *testing.B) {
	data := BuildPacket(0, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParsePacket(data, 0)
	}
}

func BenchmarkHandBuild_256B_2ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BuildPacket(2, 256)
	}
}
func BenchmarkHandParse_256B_2ext(b *testing.B) {
	data := BuildPacket(2, 256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParsePacket(data, 2)
	}
}

func BenchmarkHandBuild_1024B_4ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BuildPacket(4, 1024)
	}
}
func BenchmarkHandParse_1024B_4ext(b *testing.B) {
	data := BuildPacket(4, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParsePacket(data, 4)
	}
}

func BenchmarkHandBuild_4096B_4ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BuildPacket(4, 4096)
	}
}
func BenchmarkHandParse_4096B_4ext(b *testing.B) {
	data := BuildPacket(4, 4096)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParsePacket(data, 4)
	}
}

func BenchmarkHandBuild_64K_4ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BuildPacket(4, 65536-40-32)
	}
}
func BenchmarkHandParse_64K_4ext(b *testing.B) {
	data := BuildPacket(4, 65536-40-32)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParsePacket(data, 4)
	}
}

// ============================================================
// Эксперимент Б — Варьирование сложности
// ============================================================

func BenchmarkHandBuild_1024B_Ext0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BuildPacket(0, 1024)
	}
}
func BenchmarkHandBuild_1024B_Ext1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BuildPacket(1, 1024)
	}
}
func BenchmarkHandBuild_1024B_Ext2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BuildPacket(2, 1024)
	}
}
func BenchmarkHandBuild_1024B_Ext3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BuildPacket(3, 1024)
	}
}
func BenchmarkHandBuild_1024B_Ext4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BuildPacket(4, 1024)
	}
}

func BenchmarkHandParse_1024B_Ext0(b *testing.B) {
	data := BuildPacket(0, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParsePacket(data, 0)
	}
}
func BenchmarkHandParse_1024B_Ext1(b *testing.B) {
	data := BuildPacket(1, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParsePacket(data, 1)
	}
}
func BenchmarkHandParse_1024B_Ext2(b *testing.B) {
	data := BuildPacket(2, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParsePacket(data, 2)
	}
}
func BenchmarkHandParse_1024B_Ext3(b *testing.B) {
	data := BuildPacket(3, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParsePacket(data, 3)
	}
}
func BenchmarkHandParse_1024B_Ext4(b *testing.B) {
	data := BuildPacket(4, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParsePacket(data, 4)
	}
}
