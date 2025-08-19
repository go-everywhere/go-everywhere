import theme
from nicegui import ui
from models.user import User
from typing import List


@ui.page('/')
async def page():
    with theme.frame('Generate'):
        with ui.column().classes('mx-auto'):
            with ui.row().classes('w-full items-center px-4'):
                ui.button('Generate')