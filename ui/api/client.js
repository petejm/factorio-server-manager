import Axios from "axios";

const client = Axios.create({
    withCredentials: true,
    headers: {
        'Content-Type': 'application/json'
    }
});

// CSRF token management
let csrfToken = null;

// Fetch CSRF token from server
async function fetchCSRFToken() {
    try {
        const response = await Axios.get('/api/csrf-token', { withCredentials: true });
        csrfToken = response.data.token;
        return csrfToken;
    } catch (err) {
        console.error('Failed to fetch CSRF token:', err);
        return null;
    }
}

// Initialize CSRF token on load
fetchCSRFToken();

// Request interceptor to add CSRF token to mutating requests
client.interceptors.request.use(async config => {
    const mutatingMethods = ['POST', 'PUT', 'DELETE', 'PATCH'];
    if (mutatingMethods.includes(config.method?.toUpperCase())) {
        // Ensure we have a CSRF token
        if (!csrfToken) {
            await fetchCSRFToken();
        }
        if (csrfToken) {
            config.headers['X-CSRF-Token'] = csrfToken;
        }
    }
    return config;
}, err => Promise.reject(err));

// Response interceptor to handle errors and refresh CSRF token on 403
client.interceptors.response.use(res => res, async err => {
    const originalRequest = err.config;

    // If we get a 403 and haven't retried yet, try refreshing CSRF token
    if (err.response?.status === 403 && !originalRequest._csrfRetry) {
        originalRequest._csrfRetry = true;
        await fetchCSRFToken();
        if (csrfToken) {
            originalRequest.headers['X-CSRF-Token'] = csrfToken;
            return client(originalRequest);
        }
    }

    if (err.response?.status === 502) {
        window.flash("Service not available", "red");
    } else if (err.response?.status !== 401) {
        window.flash(err.response?.data || 'An error occurred', "red");
    }
    return Promise.reject(err);
});

export default client;