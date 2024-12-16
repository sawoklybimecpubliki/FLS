create table Files(
                      storageID text,
                      file character varying(36),
                      fileID character varying(36) primary key
);

create table Links(
                      fileID character varying(36),
                      linkID text PRIMARY KEY,
                      numberOfVisits integer,
                      lifetime timestamp,
                      foreign key (fileID) REFERENCES Files(fileID)
);