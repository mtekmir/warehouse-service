create table if not exists products(
  id bigserial unique primary key,
  name varchar unique not null
)