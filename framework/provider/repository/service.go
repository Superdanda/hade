package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
	"github.com/pkg/errors"
	"time"
)

func RegisterRepository[T any, ID comparable](service contract.RepositoryService, key string, ormRepository interface{}) {
	container := service.GetContainer()
	cacheService := container.MustMake(contract.CacheKey).(contract.CacheService)
	configService := container.MustMake(contract.ConfigKey).(contract.Config)
	expireTime := configService.GetString("cache.repository.expire")
	if expireTime == "" {
		fmt.Println("从配置文件获取缓存失败，使用默认缓存时间6个小时")
	}
	expireTime = "6h"
	duration, err := time.ParseDuration(expireTime)
	if err != nil {
		fmt.Println("从配置文件获取缓存失败，使用默认缓存时间6个小时")
	} else {
		duration = 6 * time.Hour
	}
	genericRepository := NewHadeGenericRepository[T, ID](
		key,
		container,
		NewHadeCacheRepository[T, ID](cacheService, duration),
		ormRepository.(contract.OrmRepository[T, ID]),
	)
	service.GetGenericRepositoryMap()[key] = genericRepository
}

type HadeRepositoryService struct {
	container            framework.Container
	genericRepositoryMap map[string]interface{}
	contract.RepositoryService
}

func NewHadeRepositoryService(params ...interface{}) (interface{}, error) {
	if len(params) < 2 {
		return nil, errors.New("insufficient parameters")
	}
	container, ok := params[0].(framework.Container)
	if !ok {
		return nil, errors.New("invalid container parameter")
	}
	return &HadeRepositoryService{
		container:            container,
		genericRepositoryMap: make(map[string]interface{}),
	}, nil
}

func (h *HadeRepositoryService) GetGenericRepositoryByKey(key string) interface{} {
	return h.genericRepositoryMap[key]
}

func (h *HadeRepositoryService) GetGenericRepositoryMap() map[string]interface{} {
	return h.genericRepositoryMap
}
func (h *HadeRepositoryService) GetContainer() framework.Container {
	return h.container
}

type HadeGenericRepository[T any, ID comparable] struct {
	repositoryKey   string
	container       framework.Container
	cacheRepository *HadeCacheRepository[T, ID]
	ormRepository   contract.OrmRepository[T, ID]
	contract.GenericRepository[T, ID]
}

func NewHadeGenericRepository[T any, ID comparable](repositoryKey string, container framework.Container, cacheRepository *HadeCacheRepository[T, ID],
	ormRepository contract.OrmRepository[T, ID]) *HadeGenericRepository[T, ID] {
	return &HadeGenericRepository[T, ID]{repositoryKey: repositoryKey, container: container, cacheRepository: cacheRepository, ormRepository: ormRepository}
}

type HadeCacheRepository[T any, ID comparable] struct {
	cacheService    contract.CacheService
	cacheExpiration time.Duration
}

func NewHadeCacheRepository[T any, ID comparable](cacheService contract.CacheService, cacheExpiration time.Duration) *HadeCacheRepository[T, ID] {
	return &HadeCacheRepository[T, ID]{
		cacheService:    cacheService,
		cacheExpiration: cacheExpiration,
	}
}

func (g *HadeGenericRepository[T, ID]) Save(ctx context.Context, entity *T) error {
	ormRepository := g.ormRepository
	err := ormRepository.SaveToDB(entity)
	if err != nil {
		return err
	}
	err = g.updateCache(ctx, entity)
	if err != nil {
		return err
	}
	return nil
}

func (g *HadeGenericRepository[T, ID]) updateCache(ctx context.Context, entity *T) error {
	ormRepository := g.ormRepository
	id := ormRepository.GetPrimaryKey(entity)
	return g.cacheRepository.Cache(ctx, ormRepository.GetBaseField(), id, entity)
}

func (g *HadeGenericRepository[T, ID]) FindByID(ctx context.Context, id ID) (*T, error) {
	ormRepository := g.ormRepository
	cache, _ := g.cacheRepository.FindFromCache(ctx, ormRepository.GetBaseField(), id)
	if cache != nil {
		return cache, nil
	}
	entityFromDB, err := ormRepository.FindByIDFromDB(id)
	if err != nil {
		return nil, err
	}
	go g.updateCache(ctx, entityFromDB)
	return entityFromDB, nil
}

func (g *HadeGenericRepository[T, ID]) FindByField(ctx context.Context, fieldName string, value string) ([]*T, error) {
	ormRepository := g.ormRepository
	// 获取字段查询函数
	queryFunc, ok := ormRepository.GetFieldQueryFunc(fieldName)
	if !ok {
		return nil, fmt.Errorf("no query function found for field: %s", fieldName)
	}

	// 从缓存中获取 IDs
	ids, err := g.cacheRepository.FindIDsFromCache(ctx, ormRepository.GetBaseField(), fieldName, value)
	if err != nil || ids == nil || len(ids) == 0 {
		// 缓存未命中，从数据库查询
		entitiesFromDB, err := queryFunc(value)
		if err != nil {
			return nil, err
		}

		if len(entitiesFromDB) == 0 {
			return nil, nil // 数据库中没有数据
		}

		// 更新缓存
		var idsToCache []ID
		for _, entity := range entitiesFromDB {
			id := ormRepository.GetPrimaryKey(entity)
			idsToCache = append(idsToCache, id)
			// 异步更新实体缓存
			go g.cacheRepository.Cache(ctx, ormRepository.GetBaseField(), id, entity)
		}

		// 缓存字段到 IDs 的映射
		fieldValuesToIDs := map[string][]ID{value: idsToCache}
		go g.cacheRepository.CacheFieldToIDs(ctx, ormRepository.GetBaseField(), fieldName, fieldValuesToIDs)

		return entitiesFromDB, nil
	}

	// 从缓存中获取实体
	entities, err := g.cacheRepository.FindBatchByIds(ctx, ormRepository.GetBaseField(), ids)
	if err != nil {
		return nil, err
	}

	// 检查缓存中是否有缺失的实体
	var missingIDs []ID
	for i, entity := range entities {
		if entity == nil {
			missingIDs = append(missingIDs, ids[i])
		}
	}

	if len(missingIDs) > 0 {
		// 从数据库获取缺失的实体
		missingEntities, err := ormRepository.FindByIDsFromDB(missingIDs)
		if err != nil {
			return nil, err
		}

		// 更新缓存并合并结果
		for _, entity := range missingEntities {
			entities = append(entities, entity)
			id := ormRepository.GetPrimaryKey(entity)
			go g.cacheRepository.Cache(ctx, ormRepository.GetBaseField(), id, entity)
		}
	}

	return entities, nil
}

func (g *HadeGenericRepository[T, ID]) FindByIDs(ctx context.Context, ids []ID) ([]*T, error) {
	ormRepository := g.ormRepository
	// 从缓存中获取实体
	entities, err := g.cacheRepository.FindBatchByIds(ctx, ormRepository.GetBaseField(), ids)
	if err != nil {
		return nil, err
	}

	// 记录缓存中缺失的 IDs
	var missingIDs []ID
	for i, entity := range entities {
		if entity == nil {
			missingIDs = append(missingIDs, ids[i])
		}
	}

	// 从数据库获取缺失的实体
	if len(missingIDs) > 0 {
		missingEntities, err := ormRepository.FindByIDsFromDB(missingIDs)
		if err != nil {
			return nil, err
		}

		// 更新缓存并合并结果
		for _, entity := range missingEntities {
			entities = append(entities, entity)
			id := ormRepository.GetPrimaryKey(entity)
			go g.cacheRepository.Cache(ctx, ormRepository.GetBaseField(), id, entity)
		}
	}

	// 过滤掉 nil 值的实体
	var result []*T
	for _, entity := range entities {
		if entity != nil {
			result = append(result, entity)
		}
	}

	return result, nil
}

func (g *HadeGenericRepository[T, ID]) FindByFieldIn(ctx context.Context, fieldName string, values []string) ([]*T, error) {
	ormRepository := g.ormRepository
	// 获取字段查询函数
	queryFunc, ok := ormRepository.GetFieldInQueryFunc(fieldName)
	if !ok {
		return nil, fmt.Errorf("no query function found for field: %s", fieldName)
	}

	var allEntities []*T
	var allIDs []ID
	var missingFieldValues []string

	// 对每个字段值，尝试从缓存中获取 IDs
	fieldValuesToIDs := make(map[string][]ID)
	for _, value := range values {
		ids, err := g.cacheRepository.FindIDsFromCache(ctx, ormRepository.GetBaseField(), fieldName, value)
		if err != nil || ids == nil || len(ids) == 0 {
			// 缓存未命中，记录缺失的字段值
			missingFieldValues = append(missingFieldValues, value)
		} else {
			fieldValuesToIDs[value] = ids
			allIDs = append(allIDs, ids...)
		}
	}

	// 从缓存中获取实体
	entities, err := g.cacheRepository.FindBatchByIds(ctx, ormRepository.GetBaseField(), allIDs)
	if err != nil {
		return nil, err
	}

	// 记录缓存中缺失的 IDs
	var missingIDs []ID
	idEntityMap := make(map[ID]*T)
	for i, entity := range entities {
		if entity != nil {
			allEntities = append(allEntities, entity)
			id := ormRepository.GetPrimaryKey(entity)
			idEntityMap[id] = entity
		} else {
			missingIDs = append(missingIDs, allIDs[i])
		}
	}

	// 从数据库获取缺失的实体
	if len(missingIDs) > 0 {
		missingEntities, err := ormRepository.FindByIDsFromDB(missingIDs)
		if err != nil {
			return nil, err
		}

		// 更新缓存并合并结果
		for _, entity := range missingEntities {
			allEntities = append(allEntities, entity)
			id := ormRepository.GetPrimaryKey(entity)
			idEntityMap[id] = entity
			go g.cacheRepository.Cache(ctx, ormRepository.GetBaseField(), id, entity)
		}
	}

	// 处理缓存中缺失的字段值
	if len(missingFieldValues) > 0 {
		missingEntities, err := queryFunc(missingFieldValues)
		if err != nil {
			return nil, err
		}

		// 更新缓存并合并结果
		fieldValuesToIDsToCache := make(map[string][]ID)
		for _, entity := range missingEntities {
			allEntities = append(allEntities, entity)
			id := ormRepository.GetPrimaryKey(entity)
			fieldGetter, ok := ormRepository.GetFieldValueFunc(fieldName)
			if !ok {
				continue
			}
			fieldValue := fieldGetter(entity)
			fieldValuesToIDsToCache[fieldValue] = append(fieldValuesToIDsToCache[fieldValue], id)
			go g.cacheRepository.Cache(ctx, ormRepository.GetBaseField(), id, entity)
		}

		// 更新字段到 IDs 的缓存
		go g.cacheRepository.CacheFieldToIDs(ctx, ormRepository.GetBaseField(), fieldName, fieldValuesToIDsToCache)
	}

	return allEntities, nil
}

// getKey 生成缓存键
func (r *HadeCacheRepository[T, ID]) getKey(prefix string, value any) string {
	return fmt.Sprintf("%s::%v", prefix, value)
}

func (r *HadeCacheRepository[T, ID]) getKeyWithField(prefix, fieldPrefix string, value any) string {
	return fmt.Sprintf("%s::%s::%v", prefix, fieldPrefix, value)
}

// Cache 将实体缓存到 Redis
func (r *HadeCacheRepository[T, ID]) Cache(ctx context.Context, prefix string, id ID, entity *T) error {
	key := r.getKey(prefix, id)
	return r.cacheService.SetObj(ctx, key, entity, r.cacheExpiration)
}

// CacheEvict 从缓存中删除某个实体
func (r *HadeCacheRepository[T, ID]) CacheEvict(ctx context.Context, prefix string, id ID) error {
	key := r.getKey(prefix, id)
	return r.cacheService.Del(ctx, key)
}

// FindFromCache 根据 ID 从缓存中获取实体
func (r *HadeCacheRepository[T, ID]) FindFromCache(ctx context.Context, prefix string, id ID) (*T, error) {
	key := r.getKey(prefix, id)
	var entity T
	err := r.cacheService.GetObj(ctx, key, &entity)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindBatchByIds 从缓存中批量获取实体
func (r *HadeCacheRepository[T, ID]) FindBatchByIds(ctx context.Context, prefix string, ids []ID) ([]*T, error) {
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = r.getKey(prefix, id)
	}
	keyValueMap, err := r.cacheService.GetMany(ctx, keys)
	if err != nil {
		return nil, err
	}
	results := make([]*T, len(ids))
	for i, key := range keys {
		data, ok := keyValueMap[key]
		if ok && data != "" {
			var entity T
			if err := json.Unmarshal([]byte(data), &entity); err != nil {
				return nil, err
			}
			results[i] = &entity
		} else {
			results[i] = nil
		}
	}
	return results, nil
}

// CacheFieldToID 缓存字段到 ID 的映射
func (r *HadeCacheRepository[T, ID]) CacheFieldToID(ctx context.Context, prefix, fieldPrefix, fieldValue string, id ID) error {
	key := r.getKeyWithField(prefix, fieldPrefix, fieldValue)
	return r.cacheService.SetObj(ctx, key, id, r.cacheExpiration)
}

// CacheEvictFieldToID 从缓存中删除字段到 ID 的映射
func (r *HadeCacheRepository[T, ID]) CacheEvictFieldToID(ctx context.Context, prefix, fieldPrefix, fieldValue string) error {
	key := r.getKeyWithField(prefix, fieldPrefix, fieldValue)
	return r.cacheService.Del(ctx, key)
}

// FindIDFromCache 根据字段获取对应的 ID
func (r *HadeCacheRepository[T, ID]) FindIDFromCache(ctx context.Context, prefix, fieldPrefix, fieldValue string) (ID, error) {
	key := r.getKeyWithField(prefix, fieldPrefix, fieldValue)
	var id ID
	err := r.cacheService.GetObj(ctx, key, &id)
	return id, err
}

// CacheFieldToIDs 缓存字段到多个 ID 的映射
func (r *HadeCacheRepository[T, ID]) CacheFieldToIDs(ctx context.Context, prefix, fieldPrefix string, fieldValuesToIDs map[string][]ID) error {
	data := make(map[string]string)
	for fieldValue, ids := range fieldValuesToIDs {
		key := r.getKeyWithField(prefix, fieldPrefix, fieldValue)
		idsBytes, err := json.Marshal(ids)
		if err != nil {
			return err
		}
		data[key] = string(idsBytes)
	}
	return r.cacheService.SetMany(ctx, data, r.cacheExpiration)
}

// CacheEvictFieldsToIDsBatch 从缓存中批量删除字段到 ID 的映射
func (r *HadeCacheRepository[T, ID]) CacheEvictFieldsToIDsBatch(ctx context.Context, prefix, fieldPrefix string, fieldValues []string) error {
	var keys []string
	for _, fieldValue := range fieldValues {
		keys = append(keys, r.getKeyWithField(prefix, fieldPrefix, fieldValue))
	}
	return r.cacheService.DelMany(ctx, keys)
}

// FindIDsFromCache 根据字段获取对应的 ID 列表
func (r *HadeCacheRepository[T, ID]) FindIDsFromCache(ctx context.Context, prefix, fieldPrefix, fieldValue string) ([]ID, error) {
	key := r.getKeyWithField(prefix, fieldPrefix, fieldValue)
	var ids []ID
	err := r.cacheService.GetObj(ctx, key, &ids)
	return ids, err
}
