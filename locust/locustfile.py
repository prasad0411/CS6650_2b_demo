import random
from locust import HttpUser, FastHttpUser, task, between

PRODUCT = {
    "product_id": 1,
    "sku": "ABC-123-XYZ",
    "manufacturer": "Acme Corporation",
    "category_id": 456,
    "weight": 1250,
    "some_other_id": 789,
}

class ProductUser(HttpUser):
    wait_time = between(0.1, 0.5)

    @task(3)
    def get_product(self):
        self.client.get("/products/1", name="GET /products/1")

    @task(1)
    def add_details(self):
        pid = random.randint(1, 100)
        body = dict(PRODUCT, product_id=pid)
        self.client.post(f"/products/{pid}/details", json=body, name="POST /products/:id/details")

class ProductFastUser(FastHttpUser):
    wait_time = between(0.1, 0.5)

    @task(3)
    def get_product(self):
        self.client.get("/products/1", name="GET /products/1")

    @task(1)
    def add_details(self):
        pid = random.randint(1, 100)
        body = dict(PRODUCT, product_id=pid)
        self.client.post(f"/products/{pid}/details", json=body, name="POST /products/:id/details")
