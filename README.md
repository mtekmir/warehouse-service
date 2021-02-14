# Warehouse Service 

A warehouse software that tracks the stock information of articles. Product stocks are calculated per request based on the required articles and their quantities.

## How to Run
```
docker-compose up -d
make run
```

## How to Test
```
docker-compose up -d
make test
```

## Domain 
--- 
##### Products
A product represents an end product that is made of multiple articles. 

##### Articles
An article is a part of a product. 

## Endpoints
---

### Import Products
Import products from json either by posting a json file or a json body. No return value.
##### Base URI
`/products/import`
>Example Request
```
POST /products/import HTTP/1.1
Host: localhost:8080
Accept: application/json
```
```
curl --location --request POST 'localhost:8080/products/import' \
--header 'Content-Type: application/json' \
--data-binary '<path to file>'
```
```
curl --location --request POST 'localhost:8080/products/import' \
--header 'Content-Type: application/json' \
--data-raw '{
    "products": [
        {
            "name": "Dining Chair",
            "barcode": "123",
            "contain_articles": [
                {
                    "art_id": "11",
                    "name": "leg",
                    "amount_of": "4"
                },
                {
                    "art_id": "22",
                    "name": "screw",
                    "amount_of": "8"
                }
            ]
        }
    ]
}'
```

### Get Products
Get products with stock information.
##### Base URI
`/products`
>Example Request
```
GET /products HTTP/1.1
Host: localhost:8080
Accept: application/json
```
```
curl --location --request GET 'localhost:8080/products'
```
>Example Response
```
[
    {
        "id": 11,
        "barcode": "775895845",
        "name": "Kitchen Table",
        "available_quantity": 82,
        "contain_articles": [
            {
                "art_id": "1",
                "name": "top leg",
                "stock": 963,
                "reqired_amount": 4
            },
            {
                "art_id": "2",
                "name": "big board",
                "stock": 248,
                "reqired_amount": 3
            }
        ]
    },
    {
        "id": 12,
        "barcode": "944947615",
        "name": "Drawer",
        "available_quantity": 124,
        "contain_articles": [
            {
                "art_id": "1",
                "name": "top board",
                "stock": 963,
                "reqired_amount": 2
            },
            {
                "art_id": "2",
                "name": "side board",
                "stock": 248,
                "reqired_amount": 2
            }
        ]
    }
]
```
### Get Product
Get product with stock information. 
##### Base URI
`/products/{ID}`
>Example Request
```
curl --location --request GET 'localhost:8080/products/1'
```
>Example Response
```
{
    "id": 1,
    "barcode": "123",
    "name": "Dining Chair",
    "available_quantity": 64,
    "contain_articles": [
        {
            "art_id": "11",
            "name": "side seat",
            "stock": 258,
            "reqired_amount": 4
        },
        {
            "art_id": "22",
            "name": "small board",
            "stock": 942,
            "reqired_amount": 8
        },
        {
            "art_id": "33",
            "name": "big board",
            "stock": 112,
            "reqired_amount": 1
        }
    ]
}
```


### Remove Product
Remove the required articles from the inventory. Adjusts stocks accordingly. Product stock information is returned.
##### Base URI
`/products/remove/{ID}`
```
POST /products/remove/1 HTTP/1.1
Host: localhost:8080
Accept: application/json
```

>Example Request
```
curl --location --request POST 'localhost:8080/products/remove/1' \
--header 'Content-Type: application/json' \
--data-raw '{
    "qty": 1
}'
```
>Example Response
```
{
    "id": 1,
    "barcode": "123453452",
    "name": "Dining Chair",
    "available_quantity": 2,
    "contain_articles": [
        {
            "art_id": "33",
            "name": "big screw",
            "stock": 2,
            "reqired_amount": 1
        },
        {
            "art_id": "11",
            "name": "leg",
            "stock": 12,
            "reqired_amount": 4
        },
        {
            "art_id": "22",
            "name": "screw",
            "stock": 24,
            "reqired_amount": 8
        }
    ]
}
```

### Import Articles
Import articles into the database. Returns the imported articles with current stock information. Handles duplicates.
It can import 20000 articles in 500-600ms. More than that raises a postgres error of parameter limit while querying existing articles. More items can be handled by dividing the request json body into batches, each goroutine importing 20000 articles for example.
##### Base URI
`/articles/import`
```
POST /articles/import HTTP/1.1
Host: localhost:8080
Accept: application/json
```
>Example Request
```
curl --location --request POST 'localhost:8080/articles/import' \
--header 'Content-Type: application/json' \
--data-binary '<path to file>'
```
>Example Response
```
[
    {
        "art_id": "1",
        "name": "top leg",
        "stock": 784
    },
    {
        "art_id": "2",
        "name": "big board",
        "stock": 136
    },
    {
        "art_id": "3",
        "name": "bottom leg",
        "stock": 214
    }
]
```
### Get Articles
Get all articles with stock information.
##### Base URI
`/articles`
```
POST /articles HTTP/1.1
Host: localhost:8080
Accept: application/json
```
>Example Request
```
curl --location --request GET 'localhost:8080/articles'
```
>Example Response
```
{
    "inventory": [
        {
            "art_id": "1",
            "name": "top leg",
            "stock": 963
        },
        {
            "art_id": "10",
            "name": "bottom board",
            "stock": 942
        },
        {
            "art_id": "100",
            "name": "rear leg",
            "stock": 210
        }
    ]
}
```
