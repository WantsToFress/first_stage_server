//nolint
//lint:file-ignore U1000 ignore unused code, it's generated
package model

import (
	"time"
)

type ColumnsEvent struct {
	ID, Name, Description, Start, End, CreatorID string
	Creator, CreatorRel                          string
}

type ColumnsGroup struct {
	ID, Name, Decsription string
}

type ColumnsGroupAdmin struct {
	ID, PersonID, GroupID string
	Group, Person         string
}

type ColumnsGroupEvent struct {
	ID, EventID, GroupID string
	Event, Group         string
}

type ColumnsGroupMember struct {
	ID, PersonID, GroupID string
	Group, Person         string
}

type ColumnsPerson struct {
	ID, FullName, Login, Password, RoleID string
	Role                                  string
}

type ColumnsRole struct {
	ID, Name string
}

type ColumnsSchemaMigration struct {
	ID, Dirty string
}

type ColumnsSt struct {
	Event           ColumnsEvent
	Group           ColumnsGroup
	GroupAdmin      ColumnsGroupAdmin
	GroupEvent      ColumnsGroupEvent
	GroupMember     ColumnsGroupMember
	Person          ColumnsPerson
	Role            ColumnsRole
	SchemaMigration ColumnsSchemaMigration
}

var Columns = ColumnsSt{
	Event: ColumnsEvent{
		ID:          "id",
		Name:        "name",
		Description: "description",
		Start:       "start",
		End:         "end",
		CreatorID:   "creator_id",

		Creator:    "Creator",
		CreatorRel: "CreatorRel",
	},
	Group: ColumnsGroup{
		ID:          "id",
		Name:        "name",
		Decsription: "decsription",
	},
	GroupAdmin: ColumnsGroupAdmin{
		ID:       "id",
		PersonID: "person_id",
		GroupID:  "group_id",

		Group:  "Group",
		Person: "Person",
	},
	GroupEvent: ColumnsGroupEvent{
		ID:      "id",
		EventID: "event_id",
		GroupID: "group_id",

		Event: "Event",
		Group: "Group",
	},
	GroupMember: ColumnsGroupMember{
		ID:       "id",
		PersonID: "person_id",
		GroupID:  "group_id",

		Group:  "Group",
		Person: "Person",
	},
	Person: ColumnsPerson{
		ID:       "id",
		FullName: "full_name",
		Login:    "login",
		Password: "password",
		RoleID:   "role_id",

		Role: "Role",
	},
	Role: ColumnsRole{
		ID:   "id",
		Name: "name",
	},
	SchemaMigration: ColumnsSchemaMigration{
		ID:    "version",
		Dirty: "dirty",
	},
}

type TableEvent struct {
	Name, Alias string
}

type TableGroup struct {
	Name, Alias string
}

type TableGroupAdmin struct {
	Name, Alias string
}

type TableGroupEvent struct {
	Name, Alias string
}

type TableGroupMember struct {
	Name, Alias string
}

type TablePerson struct {
	Name, Alias string
}

type TableRole struct {
	Name, Alias string
}

type TableSchemaMigration struct {
	Name, Alias string
}

type TablesSt struct {
	Event           TableEvent
	Group           TableGroup
	GroupAdmin      TableGroupAdmin
	GroupEvent      TableGroupEvent
	GroupMember     TableGroupMember
	Person          TablePerson
	Role            TableRole
	SchemaMigration TableSchemaMigration
}

var Tables = TablesSt{
	Event: TableEvent{
		Name:  "event",
		Alias: "t",
	},
	Group: TableGroup{
		Name:  "group",
		Alias: "t",
	},
	GroupAdmin: TableGroupAdmin{
		Name:  "group_admin",
		Alias: "t",
	},
	GroupEvent: TableGroupEvent{
		Name:  "group_event",
		Alias: "t",
	},
	GroupMember: TableGroupMember{
		Name:  "group_member",
		Alias: "t",
	},
	Person: TablePerson{
		Name:  "person",
		Alias: "t",
	},
	Role: TableRole{
		Name:  "role",
		Alias: "t",
	},
	SchemaMigration: TableSchemaMigration{
		Name:  "schema_migrations",
		Alias: "t",
	},
}

type Event struct {
	tableName struct{} `sql:"event,alias:t" pg:",discard_unknown_columns"`

	ID          string    `sql:"id,pk,type:uuid"`
	Name        string    `sql:"name,notnull"`
	Description *string   `sql:"description"`
	Start       time.Time `sql:"start,notnull"`
	End         time.Time `sql:"end,notnull"`
	CreatorID   string    `sql:"creator_id,type:uuid,notnull"`

	Creator    *Person `pg:"fk:creator_id"`
	CreatorRel *Person `pg:"fk:creator_id"`
}

type Group struct {
	tableName struct{} `sql:"group,alias:t" pg:",discard_unknown_columns"`

	ID          string  `sql:"id,pk,type:uuid"`
	Name        string  `sql:"name,notnull"`
	Decsription *string `sql:"decsription"`
}

type GroupAdmin struct {
	tableName struct{} `sql:"group_admin,alias:t" pg:",discard_unknown_columns"`

	ID       string `sql:"id,pk,type:uuid"`
	PersonID string `sql:"person_id,type:uuid,notnull"`
	GroupID  string `sql:"group_id,type:uuid,notnull"`

	Group  *Group  `pg:"fk:group_id"`
	Person *Person `pg:"fk:person_id"`
}

type GroupEvent struct {
	tableName struct{} `sql:"group_event,alias:t" pg:",discard_unknown_columns"`

	ID      string `sql:"id,pk,type:uuid"`
	EventID string `sql:"event_id,type:uuid,notnull"`
	GroupID string `sql:"group_id,type:uuid,notnull"`

	Event *Event `pg:"fk:event_id"`
	Group *Group `pg:"fk:group_id"`
}

type GroupMember struct {
	tableName struct{} `sql:"group_member,alias:t" pg:",discard_unknown_columns"`

	ID       string `sql:"id,pk,type:uuid"`
	PersonID string `sql:"person_id,type:uuid,notnull"`
	GroupID  string `sql:"group_id,type:uuid,notnull"`

	Group  *Group  `pg:"fk:group_id"`
	Person *Person `pg:"fk:person_id"`
}

type Person struct {
	tableName struct{} `sql:"person,alias:t" pg:",discard_unknown_columns"`

	ID       string `sql:"id,pk,type:uuid"`
	FullName string `sql:"full_name,notnull"`
	Login    string `sql:"login,notnull"`
	Password string `sql:"password,notnull"`
	RoleID   string `sql:"role_id,type:uuid,notnull"`

	Role *Role `pg:"fk:role_id"`
}

type Role struct {
	tableName struct{} `sql:"role,alias:t" pg:",discard_unknown_columns"`

	ID   string `sql:"id,pk,type:uuid"`
	Name string `sql:"name,notnull"`
}

type SchemaMigration struct {
	tableName struct{} `sql:"schema_migrations,alias:t" pg:",discard_unknown_columns"`

	ID    int64 `sql:"version,pk"`
	Dirty bool  `sql:"dirty,notnull"`
}
