create table if not exists products(
  id bigserial unique primary key,
  barcode varchar unique not null,
  name varchar not null
)