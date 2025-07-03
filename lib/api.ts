const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";

export async function startScan({ url, user_id, project_id }: {
    url: string;
    user_id: string;
    project_id: string;
}) {
    const res = await fetch(`${API_URL}/api/scan`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ url, user_id, project_id }),
    });

    if (!res.ok) {
        const error = await res.json();
        throw new Error(error.detail || "Failed to start scan");
    }

    return await res.json(); // { scan_id, status }
}
