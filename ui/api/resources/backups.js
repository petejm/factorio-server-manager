import client from "../client";

export default {
    list: async () => {
        const response = await client.get('/api/backups/list');
        return response.data;
    },
    create: async (saveName) => {
        const response = await client.post('/api/backups/create', {
            save_name: saveName
        });
        return response.data;
    },
    restore: async (id) => {
        const response = await client.post(`/api/backups/restore/${id}`);
        return response.data;
    },
    delete: async (id) => {
        const response = await client.delete(`/api/backups/${id}`);
        return response.data;
    },
    download: (id) => {
        // Return URL for direct download
        return `/api/backups/download/${id}`;
    },
    verify: async (id) => {
        const response = await client.get(`/api/backups/verify/${id}`);
        return response.data;
    }
}
