create table if not exists product_articles(
  id bigserial unique primary key,
  amount int not null,
  product_id bigint not null references products(id),
  article_id bigint not null references articles(id)
)