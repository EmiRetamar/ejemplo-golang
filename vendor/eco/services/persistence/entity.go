package persistence

/**
 * User: Santiago Vidal
 * Date: 11/05/17
 * Time: 14:01
 */

type Validable interface {
	Validate() bool
}

type Storer interface {
	Store(ent Storable) error
}

type Storable interface {
	Store(st Storer) error
}

type Entity struct {
}

func (e *Entity) Store(ent Storable, st Storer) error {
	return st.Store(ent)
}
