from app import app
import os

ADMIN_HOST = os.getenv('ADMIN_HOST', 'localhost')
ADMIN_PORT = os.getenv('ADMIN_PORT', 8080)

if __name__ == '__main__':
    app.run(host=ADMIN_HOST, port=ADMIN_PORT)
