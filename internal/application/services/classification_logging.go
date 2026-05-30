package services

import (
	"crypto/sha256"
	"encoding/hex"

	"go.uber.org/zap"
)

func classificationGoodsDescriptionFields(goodsDescription string) []zap.Field {
	hash := sha256.Sum256([]byte(goodsDescription))

	return []zap.Field{
		zap.Int("goods_description_length", len(goodsDescription)),
		zap.String("goods_description_sha256", hex.EncodeToString(hash[:])),
	}
}
