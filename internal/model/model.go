//nolint
//lint:file-ignore U1000 ignore unused code, it's generated
package model

import (
	"time"
)

type ColumnsEvent struct {
	ID, Name, Description, Start, End, CreatorID, Type, CreatedAt, UpdatedAt string
	Creator, CreatorRel                                                      string
}

type ColumnsEventMember struct {
	ID, PersonID, EventID string
	Event, Person         string
}

type ColumnsGroup struct {
	ID, Name, Description, CreatedAt, UpdatedAt string
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

type ColumnsMessage struct {
	ID, PersonID, EventID, Login, FullName, Time, Message string
}

type ColumnsPerson struct {
	ID, FullName, Login, Password, RoleID, CreatedAt, UpdatedAt string
	Role                                                        string
}

type ColumnsRole struct {
	ID, Name string
}

type ColumnsSchemaMigration struct {
	ID, Dirty string
}

type ColumnsSt struct {
	Event           ColumnsEvent
	EventMember     ColumnsEventMember
	Group           ColumnsGroup
	GroupAdmin      ColumnsGroupAdmin
	GroupEvent      ColumnsGroupEvent
	GroupMember     ColumnsGroupMember
	Message         ColumnsMessage
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
		Type:        "type",
		CreatedAt:   "created_at",
		UpdatedAt:   "updated_at",

		Creator:    "Creator",
		CreatorRel: "CreatorRel",
	},
	EventMember: ColumnsEventMember{
		ID:       "id",
		PersonID: "person_id",
		EventID:  "event_id",

		Event:  "Event",
		Person: "Person",
	},
	Group: ColumnsGroup{
		ID:          "id",
		Name:        "name",
		Description: "description",
		CreatedAt:   "created_at",
		UpdatedAt:   "updated_at",
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
	Message: ColumnsMessage{
		ID:       "id",
		PersonID: "person_id",
		EventID:  "event_id",
		Login:    "login",
		FullName: "full_name",
		Time:     "time",
		Message:  "message",
	},
	Person: ColumnsPerson{
		ID:        "id",
		FullName:  "full_name",
		Login:     "login",
		Password:  "password",
		RoleID:    "role_id",
		CreatedAt: "created_at",
		UpdatedAt: "updated_at",

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

type TableEventMember struct {
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

type TableMessage struct {
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
	EventMember     TableEventMember
	Group           TableGroup
	GroupAdmin      TableGroupAdmin
	GroupEvent      TableGroupEvent
	GroupMember     TableGroupMember
	Message         TableMessage
	Person          TablePerson
	Role            TableRole
	SchemaMigration TableSchemaMigration
}

var Tables = TablesSt{
	Event: TableEvent{
		Name:  "event",
		Alias: "t",
	},
	EventMember: TableEventMember{
		Name:  "event_member",
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
	Message: TableMessage{
		Name:  "message",
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
	Type        string    `sql:"type,notnull"`
	CreatedAt   time.Time `sql:"created_at,notnull"`
	UpdatedAt   time.Time `sql:"updated_at,notnull"`

	Creator    *Person `pg:"fk:creator_id"`
	CreatorRel *Person `pg:"fk:creator_id"`
}

type EventMember struct {
	tableName struct{} `sql:"event_member,alias:t" pg:",discard_unknown_columns"`

	ID       string `sql:"id,pk,type:uuid"`
	PersonID string `sql:"person_id,type:uuid,notnull"`
	EventID  string `sql:"event_id,type:uuid,notnull"`

	Event  *Event  `pg:"fk:event_id"`
	Person *Person `pg:"fk:person_id"`
}

type Group struct {
	tableName struct{} `sql:"group,alias:t" pg:",discard_unknown_columns"`

	ID          string    `sql:"id,pk,type:uuid"`
	Name        string    `sql:"name,notnull"`
	Description *string   `sql:"description"`
	CreatedAt   time.Time `sql:"created_at,notnull"`
	UpdatedAt   time.Time `sql:"updated_at,notnull"`
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

type Message struct {
	tableName struct{} `sql:"message,alias:t" pg:",discard_unknown_columns"`

	ID       string    `sql:"id,pk,type:uuid"`
	PersonID string    `sql:"person_id,type:uuid,notnull"`
	EventID  string    `sql:"event_id,type:uuid,notnull"`
	Login    string    `sql:"login,type:uuid,notnull"`
	FullName string    `sql:"full_name,notnull"`
	Time     time.Time `sql:"time,notnull"`
	Message  string    `sql:"message,notnull"`
}

type Person struct {
	tableName struct{} `sql:"person,alias:t" pg:",discard_unknown_columns"`

	ID        string    `sql:"id,pk,type:uuid"`
	FullName  string    `sql:"full_name,notnull"`
	Login     string    `sql:"login,notnull"`
	Password  string    `sql:"password,notnull"`
	RoleID    string    `sql:"role_id,type:uuid,notnull"`
	CreatedAt time.Time `sql:"created_at,notnull"`
	UpdatedAt time.Time `sql:"updated_at,notnull"`

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
