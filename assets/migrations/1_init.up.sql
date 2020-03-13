CREATE TABLE "event" (
                         "id" uuid NOT NULL,
                         "name" varchar(1024) NOT NULL,
                         "description" text,
                         "start" timestamp NOT NULL,
                         "end" timestamp NOT NULL,
                         "creator_id" uuid NOT NULL,
                         "type" varchar(128) not null,
                         "created_at" timestamp not null,
                         "updated_at" timestamp not null,
                         PRIMARY KEY ("id")
);

CREATE TABLE "event_member" (
                                "id" uuid NOT NULL,
                                "person_id" uuid NOT NULL,
                                "event_id" uuid NOT NULL,
                                PRIMARY KEY ("id")
);
CREATE INDEX "event_person_index" ON event_member USING btree (
                                                               "person_id"
    );
CREATE INDEX "event_event_index" ON event_member USING btree (
                                                              "event_id"
    );
CREATE UNIQUE INDEX "event_to_person_index" ON event_member USING btree (
                                                                         "person_id",
                                                                         "event_id"
    );

CREATE TABLE "group" (
                         "id" uuid NOT NULL,
                         "name" varchar(2048) NOT NULL,
                         "description" text,
                         "created_at" timestamp not null,
                         "updated_at" timestamp not null,
                         PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "group_name_index" ON "group" USING btree (
                                                               "name"
    );

CREATE TABLE "group_admin" (
                               "id" uuid NOT NULL,
                               "person_id" uuid NOT NULL,
                               "group_id" uuid NOT NULL,
                               PRIMARY KEY ("id")
);
CREATE INDEX "group_admin_to_admin_index" ON group_admin USING btree (
                                                                      "person_id"
    );
CREATE INDEX "group_admin_to_group_index" ON group_admin USING btree (
                                                                      "group_id"
    );
CREATE UNIQUE INDEX "group_admins_index" ON group_admin USING btree (
                                                                     "group_id",
                                                                     "person_id"
    );

CREATE TABLE "group_event" (
                               "id" uuid NOT NULL,
                               "event_id" uuid NOT NULL,
                               "group_id" uuid NOT NULL,
                               PRIMARY KEY ("id")
);
CREATE INDEX "group_event_event_index" ON group_event USING btree (
                                                                   "event_id"
    );
CREATE INDEX "group_event_to_group_index" ON group_event USING btree (
                                                                      "group_id"
    );
CREATE UNIQUE INDEX "group_events_index" ON group_event USING btree (
                                                                     "event_id",
                                                                     "group_id"
    );

CREATE TABLE "group_member" (
                                "id" uuid NOT NULL,
                                "person_id" uuid NOT NULL,
                                "group_id" uuid NOT NULL,
                                PRIMARY KEY ("id")
);
CREATE INDEX "group_member_to_member_index" ON group_member USING btree (
                                                                         "person_id"
    );
CREATE INDEX "group_member_to_group_index" ON group_member USING btree (
                                                                        "group_id"
    );
CREATE UNIQUE INDEX "group_members_index" ON group_member USING btree (
                                                                       "person_id",
                                                                       "group_id"
    );

CREATE TABLE "person" (
                          "id" uuid NOT NULL,
                          "full_name" varchar(2048) NOT NULL,
                          "login" varchar(256) NOT NULL,
                          "password" varchar(256) NOT NULL,
                          "role_id" uuid NOT NULL,
                          "created_at" timestamp not null,
                          "updated_at" timestamp not null,
                          PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "login_index" ON person USING btree (
                                                         "login"
    );
CREATE UNIQUE INDEX "login_password_index" ON person USING btree (
                                                                  "login",
                                                                  "password"
    );
CREATE INDEX "full_name_index" ON person USING btree (
                                                             "full_name"
    );

CREATE TABLE "role" (
                        "id" uuid NOT NULL,
                        "name" varchar(256) NOT NULL,
                        PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "role_name_index" ON role USING btree (
                                                           "name"
    );

ALTER TABLE "event" ADD CONSTRAINT "creator" FOREIGN KEY ("creator_id") REFERENCES "person" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "event" ADD CONSTRAINT "fk_event_person_1" FOREIGN KEY ("creator_id") REFERENCES "person" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;
DROP INDEX "event_person_index";
DROP INDEX "event_event_index";
DROP INDEX "event_to_person_index";
ALTER TABLE "event_member" ADD CONSTRAINT "fk_event_member_person_1" FOREIGN KEY ("person_id") REFERENCES "person" ("id");
ALTER TABLE "event_member" ADD CONSTRAINT "fk_event_member_event_1" FOREIGN KEY ("event_id") REFERENCES "event" ("id");
CREATE INDEX "event_person_index" ON event_member USING btree (
                                                               "person_id"
    );
CREATE INDEX "event_event_index" ON event_member USING btree (
                                                              "event_id"
    );
CREATE UNIQUE INDEX "event_to_person_index" ON event_member USING btree (
                                                                         "person_id",
                                                                         "event_id"
    );
DROP INDEX "group_name_index";
CREATE UNIQUE INDEX "group_name_index" ON "group" USING btree (
                                                               "name"
    );
DROP INDEX "group_admin_to_admin_index";
DROP INDEX "group_admin_to_group_index";
DROP INDEX "group_admins_index";
ALTER TABLE "group_admin" ADD CONSTRAINT "fk_group_admin_person_1" FOREIGN KEY ("person_id") REFERENCES "person" ("id");
ALTER TABLE "group_admin" ADD CONSTRAINT "fk_group_admin_group_1" FOREIGN KEY ("group_id") REFERENCES "group" ("id");
CREATE INDEX "group_admin_to_admin_index" ON group_admin USING btree (
                                                                      "person_id"
    );
CREATE INDEX "group_admin_to_group_index" ON group_admin USING btree (
                                                                      "group_id"
    );
CREATE UNIQUE INDEX "group_admins_index" ON group_admin USING btree (
                                                                     "group_id",
                                                                     "person_id"
    );
DROP INDEX "group_event_event_index";
DROP INDEX "group_event_to_group_index";
DROP INDEX "group_events_index";
ALTER TABLE "group_event" ADD CONSTRAINT "fk_group_event_event_1" FOREIGN KEY ("event_id") REFERENCES "event" ("id");
ALTER TABLE "group_event" ADD CONSTRAINT "fk_group_event_group_1" FOREIGN KEY ("group_id") REFERENCES "group" ("id");
CREATE INDEX "group_event_event_index" ON group_event USING btree (
                                                                   "event_id"
    );
CREATE INDEX "group_event_to_group_index" ON group_event USING btree (
                                                                      "group_id"
    );
CREATE UNIQUE INDEX "group_events_index" ON group_event USING btree (
                                                                     "event_id",
                                                                     "group_id"
    );
DROP INDEX "group_member_to_member_index";
DROP INDEX "group_member_to_group_index";
DROP INDEX "group_members_index";
ALTER TABLE "group_member" ADD CONSTRAINT "group_member_to_person" FOREIGN KEY ("person_id") REFERENCES "person" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "group_member" ADD CONSTRAINT "group_member_to_group" FOREIGN KEY ("group_id") REFERENCES "group" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
CREATE INDEX "group_member_to_member_index" ON group_member USING btree (
                                                                         "person_id"
    );
CREATE INDEX "group_member_to_group_index" ON group_member USING btree (
                                                                        "group_id"
    );
CREATE UNIQUE INDEX "group_members_index" ON group_member USING btree (
                                                                       "person_id",
                                                                       "group_id"
    );
DROP INDEX "login_index";
DROP INDEX "login_password_index";
DROP INDEX "full_name_index";
ALTER TABLE "person" ADD CONSTRAINT "fk_person_role_1" FOREIGN KEY ("role_id") REFERENCES "role" ("id");
CREATE UNIQUE INDEX "login_index" ON person USING btree (
                                                         "login"
    );
CREATE UNIQUE INDEX "login_password_index" ON person USING btree (
                                                                  "login",
                                                                  "password"
    );
CREATE INDEX "full_name_index" ON person USING btree (
                                                             "full_name"
    );
DROP INDEX "role_name_index";
CREATE UNIQUE INDEX "role_name_index" ON role USING btree (
                                                           "name"
    );

CREATE TABLE "chat" (
                        "id" uuid NOT NULL,
                        "event_id" uuid not null,
                        "login" uuid not null,
                        "time" int8,
                        PRIMARY KEY ("id")
);

CREATE INDEX "chat_login_index" ON chat USING btree (
                                                                  "login"
    );
CREATE INDEX "chat_event_index" ON chat USING btree (
                                                                  "event_id"
    );
