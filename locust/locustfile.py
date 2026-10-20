import random
from locust import FastHttpUser, task, between

TERMS = ["Electro", "Alpha", "Beta", "Books", "Home", "Gamma", "Sports", "Delta"]

class SearchUser(FastHttpUser):
    wait_time = between(0.0, 0.1)

    @task
    def search(self):
        q = random.choice(TERMS)
        self.client.get(f"/products/search?q={q}", name="GET /products/search")
