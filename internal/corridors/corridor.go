package corridors

import (
	"sort"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type CorridorPriority int

const (
	FromCategory = CorridorPriority(iota)
	FromTag
	FromNode
	ToCategory
	ToTag
	ToNode
	Channel
)

func Category() CorridorType {
	return CorridorType{0, "category", 0}
}
func Tag() CorridorType {
	return CorridorType{1, "tag", 0}
}
func corridorTypes() []CorridorType {
	return []CorridorType{Category(), Tag()}
}

type CorridorType struct {
	CorridorTypeId int
	Label          string
	DefaultFlag    int
}

type CorridorKey struct {
	CorridorType   CorridorType `json:"corridorType"`
	Inverse        bool         `json:"inverse"`
	ReferenceId    int          `json:"referenceId"`
	FromCategoryId int          `json:"fromCategoryId"`
	FromTagId      int          `json:"fromTagId"`
	FromNodeId     int          `json:"fromNodeId"`
	ChannelId      int          `json:"channelId"`
	ToCategoryId   int          `json:"toCategoryId"`
	ToTagId        int          `json:"toTagId"`
	ToNodeId       int          `json:"toNodeId"`
}

type Corridor struct {
	CorridorId     int       `json:"corridorId" db:"corridor_id"`
	CorridorTypeId int       `json:"corridorTypeId" db:"corridor_type_id"`
	ReferenceId    *int      `json:"referenceId" db:"reference_id"`
	Flag           int       `json:"flag" db:"flag"`
	Inverse        bool      `json:"inverse" db:"inverse"`
	Priority       int       `json:"priority" db:"priority"`
	FromCategoryId *int      `json:"fromCategoryId" db:"from_category_id"`
	FromTagId      *int      `json:"fromTagId" db:"from_tag_id"`
	FromNodeId     *int      `json:"fromNodeId" db:"from_node_id"`
	ChannelId      *int      `json:"channelId" db:"channel_id"`
	ToCategoryId   *int      `json:"toCategoryId" db:"to_category_id"`
	ToTagId        *int      `json:"toTagId" db:"to_tag_id"`
	ToNodeId       *int      `json:"toNodeId" db:"to_node_id"`
	CreatedOn      time.Time `json:"createdOn" db:"created_on"`
	UpdateOn       time.Time `json:"updatedOn" db:"updated_on"`
}

type CorridorNodesChannels struct {
	Corridors     []*CorridorFields `json:"corridors"`
	TotalNodes    int               `json:"totalNodes" db:"total_nodes"`
	TotalChannels int               `json:"totalChannels" db:"total_channels"`
}

type CorridorFields struct {
	CorridorId     int     `json:"corridorId" db:"corridor_id"`
	ReferenceId    *int    `json:"referenceId" db:"reference_id"`
	Alias          *string `json:"alias" db:"alias"`
	ShortChannelId *string `json:"shortChannelId" db:"short_channel_id"`
}

type corridorCacheByType struct {
	corridorCacheLock       sync.RWMutex
	corridorCacheMap        map[int]map[CorridorKey]Corridor
	corridorCacheSortedKeys []int
}

func (cc *corridorCacheByType) updateCache(cacheMap map[int]map[CorridorKey]Corridor, cacheSortedKeys []int) {
	cc.corridorCacheLock.Lock()
	defer cc.corridorCacheLock.Unlock()
	cc.corridorCacheMap = cacheMap
	cc.corridorCacheSortedKeys = cacheSortedKeys
}

func (cc *corridorCacheByType) getBestCorridor(key CorridorKey) Corridor {
	cc.corridorCacheLock.RLock()
	defer cc.corridorCacheLock.RUnlock()

	for _, priority := range cc.corridorCacheSortedKeys {
		for cKey, c := range cc.corridorCacheMap[priority] {
			if c.ReferenceId != nil && key.ReferenceId == *c.ReferenceId && equals(key, priority, cKey) {
				return c
			}
		}
	}
	return Corridor{CorridorTypeId: key.CorridorType.CorridorTypeId, Flag: key.CorridorType.DefaultFlag}
}

var corridorCache = map[CorridorType]*corridorCacheByType{ //nolint:gochecknoglobals
	Category(): {
		sync.RWMutex{},
		make(map[int]map[CorridorKey]Corridor, 0),
		[]int{},
	},
	Tag(): {
		sync.RWMutex{},
		make(map[int]map[CorridorKey]Corridor, 0),
		[]int{},
	},
}

func finalizeCorridorCacheByType(corridorType CorridorType, corridorStagingCache *map[int]map[CorridorKey]Corridor) {
	var corridorCacheSortedKeys []int
	for k := range *corridorStagingCache {
		corridorCacheSortedKeys = append(corridorCacheSortedKeys, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(corridorCacheSortedKeys)))
	corridorCache[corridorType].updateCache(*corridorStagingCache, corridorCacheSortedKeys)
}

func RefreshCorridorCache(db *sqlx.DB) error {
	for _, corridorType := range corridorTypes() {
		err := RefreshCorridorCacheByType(db, corridorType)
		if err != nil {
			return err
		}
	}
	return nil
}

func RefreshCorridorCacheByTypeId(db *sqlx.DB, corridorTypeId int) error {
	return RefreshCorridorCacheByType(db, *getCorridorTypeFromId(corridorTypeId))
}

func RefreshCorridorCacheByType(db *sqlx.DB, corridorType CorridorType) error {
	corridorStagingCache := make(map[int]map[CorridorKey]Corridor, 0)
	corridors, err := getCorridorsByCorridorType(db, corridorType)
	if err != nil {
		return errors.Wrapf(err, "Obtaining corridors. (%v %v)", corridorType, db)
	}
	for _, c := range corridors {
		addToCorridorCache(*c, &corridorStagingCache)
	}
	finalizeCorridorCacheByType(corridorType, &corridorStagingCache)
	return nil
}

func addToCorridorCache(c Corridor, corridorStagingCache *map[int]map[CorridorKey]Corridor) {
	priority := calculatePriority(c)
	if c.Priority != priority {
		log.Error().Msgf("Priority mismatch for corridorId: %v", c.CorridorId)
	}
	if c.Inverse {
		log.Error().Msgf("Inverse corridors are not implemented yet corridorId: %v", c.CorridorId)
	} else {
		if corridorStagingCache == nil {
			newMap := make(map[int]map[CorridorKey]Corridor)
			corridorStagingCache = &newMap
		}
		if (*corridorStagingCache)[priority] == nil {
			(*corridorStagingCache)[priority] = make(map[CorridorKey]Corridor)
		}
		(*corridorStagingCache)[priority][constructKey(c)] = c
	}
}

func getCorridorTypeFromId(corridorTypeId int) *CorridorType {
	switch corridorTypeId {
	case Category().CorridorTypeId:
		category := Category()
		return &category
	case Tag().CorridorTypeId:
		tag := Tag()
		return &tag
	}
	return nil
}

func constructKey(corridor Corridor) CorridorKey {
	key := CorridorKey{}
	corridorType := getCorridorTypeFromId(corridor.CorridorTypeId)
	if corridorType == nil {
		return key
	}
	key.CorridorType = *corridorType
	if corridor.ReferenceId != nil {
		key.ReferenceId = *corridor.ReferenceId
	}
	key.Inverse = corridor.Inverse
	if corridor.FromCategoryId != nil {
		key.FromCategoryId = *corridor.FromCategoryId
	}
	if corridor.FromTagId != nil {
		key.FromTagId = *corridor.FromTagId
	}
	if corridor.FromNodeId != nil {
		key.FromNodeId = *corridor.FromNodeId
	}
	if corridor.ToCategoryId != nil {
		key.ToCategoryId = *corridor.ToCategoryId
	}
	if corridor.ToTagId != nil {
		key.ToTagId = *corridor.ToTagId
	}
	if corridor.ToNodeId != nil {
		key.ToNodeId = *corridor.ToNodeId
	}
	if corridor.ChannelId != nil {
		key.ChannelId = *corridor.ChannelId
	}
	return key
}

func calculatePriority(corridor Corridor) int {
	priority := 0
	if corridor.FromCategoryId != nil && *corridor.FromCategoryId != 0 {
		priority += getPriority(FromCategory)
	}
	if corridor.FromTagId != nil && *corridor.FromTagId != 0 {
		priority += getPriority(FromTag)
	}
	if corridor.FromNodeId != nil && *corridor.FromNodeId != 0 {
		priority += getPriority(FromNode)
	}
	if corridor.ToCategoryId != nil && *corridor.ToCategoryId != 0 {
		priority += getPriority(ToCategory)
	}
	if corridor.ToTagId != nil && *corridor.ToTagId != 0 {
		priority += getPriority(ToTag)
	}
	if corridor.ToNodeId != nil && *corridor.ToNodeId != 0 {
		priority += getPriority(ToNode)
	}
	if corridor.ChannelId != nil && *corridor.ChannelId != 0 {
		priority += getPriority(Channel)
	}
	return priority
}

func getPriority(corridorPriority CorridorPriority) int {
	return 1 << corridorPriority
}

func hasPriority(corridorPriority CorridorPriority, priority int) bool {
	calculatedPriority := getPriority(corridorPriority)
	return (priority & calculatedPriority) == calculatedPriority
}

func equals(key CorridorKey, priority int, otherKey CorridorKey) bool {
	if hasPriority(FromCategory, priority) && otherKey.FromCategoryId != key.FromCategoryId {
		return false
	}
	if hasPriority(FromTag, priority) && otherKey.FromTagId != key.FromTagId {
		return false
	}
	if hasPriority(FromNode, priority) && otherKey.FromNodeId != key.FromNodeId {
		return false
	}
	if hasPriority(ToCategory, priority) && otherKey.ToCategoryId != key.ToCategoryId {
		return false
	}
	if hasPriority(ToTag, priority) && otherKey.ToTagId != key.ToTagId {
		return false
	}
	if hasPriority(ToNode, priority) && otherKey.ToNodeId != key.ToNodeId {
		return false
	}
	if hasPriority(Channel, priority) && otherKey.ChannelId != key.ChannelId {
		return false
	}
	return true
}

func GetBestCorridor(key CorridorKey) Corridor {
	return corridorCache[key.CorridorType].getBestCorridor(key)
}

func GetBestCorridorFlag(key CorridorKey) int {
	corridor := GetBestCorridor(key)
	return corridor.Flag
}

func GetBestCorridorStatus(key CorridorKey) bool {
	corridor := GetBestCorridor(key)
	return corridor.Flag == 1
}
