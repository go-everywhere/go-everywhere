from contextlib import contextmanager
from nicegui import ui


@contextmanager
def frame(navigation_title: str):
    ui.colors(primary='#53B689', secondary='#6E93D6', accent='#111B1E', positive='#53B689')
    with ui.header():
        ui.label('Assette').classes('font-bold')
        ui.space()
        ui.link('My models', '/models')
        ui.link('Generate', '/')
        ui.link('Profile', '/profile')
        ui.space()
    with ui.column().classes('absolute-center items-center'):
        yield