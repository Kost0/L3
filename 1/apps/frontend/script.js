const API_BASE_URL = window.API_URL

export async function getStatus(id) {
    const response = await fetch(`${API_BASE_URL}/notify/${encodeURIComponent(id)}`, {
        method: 'GET',
        headers: {
            'Accept': 'application/json'
        },
    });

    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Ошибка: ${response.status} ${errorText}`);
    }

    return await response.json();
}

export async function sendNotification(data) {
    const response = await fetch(`${API_BASE_URL}/notify`, {
        method: 'POST',
        headers: {
            'Content-type': 'application/json',
        },
        body: JSON.stringfy(data)
    });

    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Ошибка: ${response.status} ${errorText}`);
    }

    return await response.json();
}

export async function deleteNotification(id) {
    const response = await fetch(`${API_BASE_URL}/notify/${encodeURIComponent(id)}`, {
        method: 'DELETE',
    });

    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Ошибка: ${response.status} ${errorText}`);
    }
}