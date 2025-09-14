import requests

BASE_URL = "https://www.hut-reservation.org"
USERNAME = "test@test.net"
PASSWORD = "password"

# persistent session
session = requests.Session()
session.headers.update({
    "User-Agent": "Mozilla/5.0",
    "Accept": "application/json, text/plain, */*",
    "Referer": f"{BASE_URL}/login"
})

# retrieve CSRF token
csrf_url = f"{BASE_URL}/api/v1/csrf"
r = session.get(csrf_url)
r.raise_for_status()

xsrf_token = session.cookies.get("XSRF-TOKEN")
print("CSRF token initial:", xsrf_token)

# login
login_url = f"{BASE_URL}/api/v1/users/login"
login_data = {
    "username": USERNAME,
    "password": PASSWORD
}
r = session.post(
    login_url,
    data=login_data,
    headers={"X-XSRF-TOKEN": xsrf_token}
)
r.raise_for_status()

# (optional) check current user
me_url = f"{BASE_URL}/api/v1/manage/currentUser"
r = session.get(me_url)
print("Current user:", r.json())

# list huts
huts_url = f"{BASE_URL}/api/v1/manage/hutsList"
r = session.get(huts_url)
print("Huts list:", r.json())
