package bitfield

// bitmap
type BitField []byte

// 判断是否存在piece
func (bf BitField) HasPiece(idx int) bool {

	// peice位于整个数组中的下标
	byteIdx := idx / 8

	if len(bf) < byteIdx {
		return false
	}

	// piece在对应下标的位置
	offset := idx % 8

	// 每个字节分别存储8个piece
	// 对应字节位位1代表存在piece
	// 右移7-offset正好可以得到对应piece的位
	// 1 & 1 = 1  0 & 1 = 0
	return bf[byteIdx]>>uint(7-offset)&1 != 0
}

// 设置对应字节位上的piece存在
func (bf BitField) SetPiece(idx int) {

	byteIdx := idx / 8

	if byteIdx < 0 || byteIdx >= len(bf) {
		return
	}

	offset := idx % 8

	// 相反的设置对应piece的位
	// 则是左移相或
	// 1 | 1 = 1  1 | 0 = 1
	bf[byteIdx] |= 1 << uint(7-offset)
}
