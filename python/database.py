from tortoise import Tortoise
from nicegui import app
import models.user

async def init_db() -> None:
    await Tortoise.init(db_url='sqlite://db.sqlite3', modules={'models': [
        'models.user'
    ]})
    await Tortoise.generate_schemas()


async def close_db() -> None:
    await Tortoise.close_connections()

app.on_startup(init_db)
app.on_shutdown(close_db)