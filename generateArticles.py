import sys
import json
from random import randrange

prod_names = [
    "chair",
    "towel",
    "wardrobe",
    "bed",
    "pillow",
    "table",
    "lamp",
    "mirror",
    "carpet",
]
art_names = [
    "screw",
    "seat",
    "leg",
    "screwdriver",
    "board",
    "chipboard",
    "door",
    "rail",
]
adjs = ["top", "bottom", "side", "rear", "big", "small"]


def gen_articles(n):
    arts = []
    for i in range(n):
        arts.append(
            {
                "art_id": str(i + 1),
                "name": f"{adjs[randrange(len(adjs))]} {art_names[randrange(len(art_names))]}",
                "stock": str(randrange(500)),
            }
        )

    return arts


def gen_products(n):
    prods = []
    for _ in range(n):
        prods.append(
            {
                "name": f"{adjs[randrange(len(adjs))]} {prod_names[randrange(len(prod_names))]}",
                "barcode": "{}".format(randrange(100000000, 999999999)),
                "contain_articles": gen_articles(randrange(1, 3)),
            }
        )
    return prods

def gen(n):
    inv = {"inventory": gen_articles(n)}
    with open("articles.json", "w") as f:
        json.dump(inv, f)

    prods = {"products": gen_products(n)}
    with open("products.json", "w") as f:
        json.dump(prods, f)


if __name__ == "__main__":
    args = sys.argv
    if len(args) == 1:
        gen(5)
    else:
        print(args[1])
        gen(int(args[1]))