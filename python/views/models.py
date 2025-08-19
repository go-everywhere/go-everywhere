import theme
from nicegui import ui


@ui.page('/models')
async def page():
    with theme.frame('Models'):
        with ui.column().classes('mx-auto'):
            with ui.row().classes('w-full items-center px-4'):
                ui.button('Models')