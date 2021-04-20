package cache

import lru "github.com/hashicorp/golang-lru"

var (
	QQ2TGCache *lru.Cache
	TG2QQCache *lru.Cache
	QQMID2MSG  *lru.Cache
)

func Init() {
	QQ2TGCache, _ = lru.New(200)
	TG2QQCache, _ = lru.New(200)
	QQMID2MSG, _ = lru.New(200)
}
