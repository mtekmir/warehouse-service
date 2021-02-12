create table if not exists articles(
  id bigserial unique primary key,
  art_id varchar unique not null,
  name varchar not null,
  stock int default 0
)