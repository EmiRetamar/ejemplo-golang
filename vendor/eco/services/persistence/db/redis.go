package db

import (
	"eco/services/config"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/ian-kent/go-log/log"
)

/**
 * User: Santiago Vidal
 * Date: 12/06/17
 * Time: 16:17
 *
 * Libreria de persistencia para ECO
 * Utiliza RedisDB
 */
type multireader struct {
	values []interface{}
	err    error
}

func (mr *multireader) Scan(dest ...interface{}) error {
	if mr.err != nil {
		return mr.err
	}
	_, err := redis.Scan(mr.values, dest)
	return err
}

type ecoRedisSession struct {
	conn redis.Conn
}

func (s *ecoRedisSession) GetConn() redis.Conn { //TODO: QUITAR!!!
	return s.conn
}
func (s *ecoRedisSession) Ping() (err error) {
	_, err = s.conn.Do("PING")
	return
}
func (s *ecoRedisSession) SendRaw(cmd string) ([]string, error) {

	args := strings.Split(cmd, " ")
	fields := make([]interface{}, len(args)-1)
	for i := 0; i < len(args)-1; i++ {
		fields[i] = args[i+1]
	}

	reply, err := s.conn.Do(args[0], fields...)
	if err != nil {
		return nil, err
	}

	switch reply := reply.(type) {
	case []interface{}:

		reply, err := redis.Values(s.conn.Do(args[0], fields...))
		if err != nil {
			return nil, err
		}

		list := make([]string, len(reply))
		for i := 0; i < len(reply); i++ {
			list[i] = getStringValueFromRedisReply(reply, i)
		}
		return list, nil

	default:
		return nil, nil
	}
}
func (s *ecoRedisSession) ReadEntity(key string, out interface{}) error {

	var data string
	reply, err := redis.Values(s.conn.Do("HMGET", key, "entity"))
	if err != nil {
		return err
	}

	if _, err := redis.Scan(reply, &data); err != nil {
		return err
	}
	json.NewDecoder(strings.NewReader(data)).Decode(out)
	return nil
}
func (s *ecoRedisSession) MultiRead(key string, args ...interface{}) *multireader {

	reply, err := redis.Values(s.conn.Do("HMGET", key, args))
	if err != nil {
		return &multireader{
			values: nil,
			err:    err,
		}
	}
	return &multireader{
		values: reply,
	}
}
func (s *ecoRedisSession) Store(key string, value RedisAware) error {

	ae, ok := value.(AuditEntity)
	if ok { //audita?

		ai, user := ae.AuditInfo()
		if len(user) > 0 {
			if b, _ := s.Exists(key); b {

				var oldAudit AuditData
				if err := s.ReadEntity(key, &oldAudit); err != nil {
					return err
				}
				ai.CreatedBy = oldAudit.CreatedBy
				ai.CreatedAt = oldAudit.CreatedAt
				ai.ModifiedAt = time.Now()
				ai.ModifiedBy = user
			} else {
				ai.CreatedAt = time.Now()
				ai.CreatedBy = user
			}
		}
	}

	fields := make([]interface{}, 0)
	fields = append(fields, key)
	value.RedisFillFields(value, &fields)
	_, err := s.conn.Do("HMSET", fields...)

	return err
}
func (s *ecoRedisSession) Remove(key string) error {
	_, err := s.conn.Do("DEL", key)
	return err
}
func (s *ecoRedisSession) ExpiresAt(key string, expiration time.Time) error {
	_, err := s.conn.Do("EXPIREAT", key, expiration.Unix())
	return err
}
func (s *ecoRedisSession) Exists(key string) (bool, error) {
	return redis.Bool(s.conn.Do("EXISTS", key))
}
func (s *ecoRedisSession) Sets(setName string) *ecoRedisSet {
	return &ecoRedisSet{
		session: s,
		setName: setName,
	}
}
func (s *ecoRedisSession) Raw(key string) *ecoRedisRaw {
	return &ecoRedisRaw{
		session: s,
		key: key,
	}
}
func (s *ecoRedisSession) Close() error {
	return s.conn.Close()
}

type ecoRedisRaw struct {
	session *ecoRedisSession
	key     string
}

func(r *ecoRedisRaw) SetString(id string, value string) {
	r.session.SendRaw("HSET " + r.key + " " + id + " " + value)
}

func(r *ecoRedisRaw) GetString(id string) string {
	raw, _ := r.session.SendRaw("HMGET " + r.key + " " + id)
	if len(raw) > 0 {
		return raw[0]
	} else {
		return ""
	}
}

func(r *ecoRedisRaw) SetStringArray(id string, value []string) {
	data, _ := json.Marshal(value)
	r.session.SendRaw("HSET " + r.key + " " + id + " " + string(data))
}

func(r *ecoRedisRaw) GetStringArray(id string) []string {

	var ret []string
	raw, _ := r.session.SendRaw("HMGET " + r.key + " " + id)
	if len(raw) > 0{
		json.NewDecoder(strings.NewReader(raw[0])).Decode(&ret)
	} else {
		ret = make([]string, 0)
	}
	return ret
}

type ecoRedisSet struct {
	session *ecoRedisSession
	setName string
}

func (set *ecoRedisSet) IsMember(key string) (bool, error) {

	exists, err := redis.Bool(set.session.conn.Do("SISMEMBER", set.setName, key))
	if err != nil {
		return false, err
	}
	return exists, nil
}
func (set *ecoRedisSet) List() ([]string, error) {

	reply, err := redis.Values(set.session.conn.Do("SMEMBERS", set.setName))
	if err != nil {
		return nil, err
	}
	list := make([]string, len(reply))
	for i := 0; i < len(reply); i++ {
		list[i] = getStringValueFromRedisReply(reply, i)
	}
	return list, err

}
func (set *ecoRedisSet) Add(key string) error {

	if _, err := set.session.conn.Do("SADD", set.setName, key); err != nil {
		return err
	}
	return nil
}

func (set *ecoRedisSet) Remove(key string) error {
	_, err := set.session.conn.Do("SREM", set.setName, key)
	return err
}

type ecoRedis struct {
	pool *redis.Pool
}

func (er *ecoRedis) Session() *ecoRedisSession {
	return &ecoRedisSession{
		conn: er.pool.Get(),
	}
}

//Redis da acceso al metodo Session() para trabajar con la base de datos
var Redis *ecoRedis

func initRedis() error {

	host, _ := config.EcoCfg.String("database.redis.host")
	port, _ := config.EcoCfg.String("database.redis.port")
	maxActive, _ := config.EcoCfg.Int("database.redis.maxActive")
	maxIdle, _ := config.EcoCfg.Int("database.redis.maxIdle")
	idleTimeout, _ := config.EcoCfg.Int("database.redis.idleTimeout")

	log.Info(fmt.Sprintf("Initializing Redis connection on %s:%s...", host, port))

	redisPool := &redis.Pool{
		MaxActive:   maxActive,
		MaxIdle:     maxIdle,
		IdleTimeout: time.Duration(idleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", host+":"+port)
		},
	}
	Redis = &ecoRedis{
		pool: redisPool,
	}

	return testConnection()
}

func testConnection() error {
	s := Redis.Session()
	defer s.Close()

	if err := s.Ping(); err != nil {
		return err
	}
	log.Info("Redis connection: OK")
	return nil
}

func getStringValueFromRedisReply(bi []interface{}, index int) string {

	val := bi[index]
	if  val != nil {
		return string(val.([]uint8))
	}
	return ""
}
