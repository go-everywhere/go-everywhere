import theme
from nicegui import ui
from models.user import User
from typing import List

@ui.refreshable
async def list_users():
    async def delete(user: User):
        await user.delete()
        list_users.refresh()

    users: List[User] = await User.all()
    for user in reversed(users):
        with ui.card():
            with ui.row().classes('items-center'):
                ui.input('Name', on_change=user.save) \
                    .bind_value(user, 'name').on('blur', list_users.refresh)
                ui.input('Email', on_change=user.save) \
                    .bind_value(user, 'email').on('blur', list_users.refresh)
                ui.button(icon='delete', on_click=lambda e, u=user: delete(u)).props('flat')

@ui.page('/profile')
async def page():
    async def create():
        await User.create(name=name.value, email=email.value or 0)
        name.value = ''
        email.value = ''
        list_users.refresh()

    with theme.frame('Profile'):
        with ui.column().classes('mx-auto'):
            with ui.row().classes('w-full items-center px-4'):
                name = ui.input(label='Name')
                email = ui.input(label='Email')
                ui.button(on_click=create, icon='add').props('flat').classes('ml-auto')
            await list_users() # type: ignore