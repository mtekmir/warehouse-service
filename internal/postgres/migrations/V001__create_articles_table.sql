create table if not exists articles(
  id bigserial unique primary key,
  barcode varchar unique not null,
  name varchar unique not null,
  stock int default 0
)