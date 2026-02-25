import { useState, type FormEvent } from 'react';
import { tenants } from '../api/client';
import { useTenant } from '../contexts/TenantContext';
import './OnboardingPage.css';

function toSlug(name: string): string {
    return name
        .toLowerCase()
        .trim()
        .replace(/[^a-z0-9]/g, '');
}

export default function OnboardingPage() {
    const { refetch } = useTenant();
    const [name, setName] = useState('');
    const [slug, setSlug] = useState('');
    const [slugTouched, setSlugTouched] = useState(false);
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);

    const handleNameChange = (val: string) => {
        setName(val);
        if (!slugTouched) {
            setSlug(toSlug(val));
        }
    };

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault();
        if (!name.trim() || !slug.trim()) return;
        setError('');
        setLoading(true);
        try {
            await tenants.create(name.trim(), slug.trim());
            await refetch();
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : 'Failed to create workspace');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="onboarding-page">
            <form className="onboarding-card" onSubmit={handleSubmit}>
                <div style={{ textAlign: 'center', marginBottom: 'var(--s-6)' }}>
                    <span style={{ fontSize: '2rem' }}>◆</span>
                </div>
                <h1>Welcome to Converda</h1>
                <p>Create your first workspace to get started. You can invite team members later.</p>

                {error && <div className="onboarding-error">{error}</div>}

                <div className="form-group">
                    <label>Workspace Name</label>
                    <input
                        type="text"
                        value={name}
                        onChange={e => handleNameChange(e.target.value)}
                        placeholder="My Company"
                        required
                        autoFocus
                    />
                </div>

                <div className="form-group">
                    <label>Slug</label>
                    <input
                        type="text"
                        value={slug}
                        onChange={e => { setSlug(e.target.value); setSlugTouched(true); }}
                        placeholder="my-company"
                        required
                    />
                    <div className="slug-preview">Your workspace URL: converda.io/<strong>{slug || '...'}</strong></div>
                </div>

                <button type="submit" className="btn btn-primary" style={{ width: '100%', marginTop: 'var(--s-4)' }} disabled={loading}>
                    {loading ? 'Creating...' : 'Create Workspace'}
                </button>
            </form>
        </div>
    );
}
