package role

type Role int

const (
	Creator Role = iota // 0 - Создатель заявок
	Manager             // 1 - Модератор
	Admin               // 2 - Администратор
)

func (r Role) String() string {
	return [...]string{"creator", "manager", "admin"}[r]
}
