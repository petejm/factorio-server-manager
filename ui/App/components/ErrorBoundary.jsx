import React from 'react';
import PropTypes from 'prop-types';

class ErrorBoundary extends React.Component {
    constructor(props) {
        super(props);
        this.state = { hasError: false, error: null, errorInfo: null };
    }

    static getDerivedStateFromError(error) {
        return { hasError: true, error };
    }

    componentDidCatch(error, errorInfo) {
        this.setState({ errorInfo });
        // You can also log the error to an error reporting service here
        if (process.env.NODE_ENV === 'development') {
            console.error('ErrorBoundary caught an error:', error, errorInfo);
        }
    }

    render() {
        if (this.state.hasError) {
            if (this.props.fallback) {
                return this.props.fallback;
            }

            return (
                <div className="p-4 bg-red-dark text-white rounded">
                    <h2 className="text-xl font-bold mb-2">Something went wrong</h2>
                    <p className="mb-4">An error occurred while rendering this component.</p>
                    {process.env.NODE_ENV === 'development' && this.state.error && (
                        <details className="bg-black bg-opacity-25 p-2 rounded">
                            <summary className="cursor-pointer">Error details</summary>
                            <pre className="mt-2 text-sm overflow-auto">
                                {this.state.error.toString()}
                                {this.state.errorInfo?.componentStack}
                            </pre>
                        </details>
                    )}
                    <button
                        className="mt-4 bg-white text-red-dark px-4 py-2 rounded hover:bg-gray-light"
                        onClick={() => this.setState({ hasError: false, error: null, errorInfo: null })}
                    >
                        Try again
                    </button>
                </div>
            );
        }

        return this.props.children;
    }
}

ErrorBoundary.propTypes = {
    children: PropTypes.node.isRequired,
    fallback: PropTypes.node
};

ErrorBoundary.defaultProps = {
    fallback: null
};

export default ErrorBoundary;
