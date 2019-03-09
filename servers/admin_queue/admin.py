from app import app
import os

ADMIN_HOST = os.getenv(ADMIN_HOST, 'localhost')

if __name__ == '__main__':
    # app.run(host=ADMIN_HOST, port=80)
    app.run()
