create table if not exists articles(
  id bigserial unique primary key,
  name varchar unique not null,
  stock int default 0
)