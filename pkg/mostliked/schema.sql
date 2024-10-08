create table if not exists posts (
  uri text primary key,
  create_ts int not null,
  likes int not null
);

create table if not exists langs (
  uri text primary key,
  lang text not null,
  foreign key(uri) references posts(uri) on delete cascade
);

create index if not exists ts_idx on posts(create_ts);
