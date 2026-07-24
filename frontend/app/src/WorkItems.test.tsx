import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen, fireEvent } from '@solidjs/testing-library';
import { WorkItems } from './WorkItems';

const envelope = (data: unknown) => ({
  ok: true,
  status: 200,
  json: async () => ({ status: true, data, error: null }),
});

const okData = (data: unknown, status = 200) => ({
  ok: true,
  status,
  json: async () => ({ status: true, data, error: null }),
});

afterEach(() => vi.unstubAllGlobals());

describe('<WorkItems>', () => {
  it('renders work items fetched from the API', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        envelope([
          {
            id: '1',
            title: 'First task',
            project_id: 'demo',
            description_markdown: '',
            priority: 1,
            labels: ['backend'],
          },
        ]),
      ),
    );

    render(() => <WorkItems />);

    expect(await screen.findByText('First task')).toBeInTheDocument();
    // 'backend' is a label chip (unique); 'Low' also appears as a <select> option.
    expect(screen.getByText('backend')).toBeInTheDocument();
    expect(screen.getAllByText('Low').length).toBeGreaterThan(0);
  });

  it('creates a work item through the API and shows it', async () => {
    const created = { id: '2', title: 'New task', project_id: 'demo', description_markdown: '' };
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(envelope([])) // initial list (empty)
      .mockResolvedValueOnce(okData(created, 201)) // POST
      .mockResolvedValueOnce(envelope([created])); // refetch
    vi.stubGlobal('fetch', fetchMock);

    render(() => <WorkItems />);
    await screen.findByText('No work items yet. Add one above.');

    fireEvent.input(screen.getByPlaceholderText('New work item title…'), {
      target: { value: 'New task' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Add' }));

    expect(await screen.findByText('New task')).toBeInTheDocument();

    const postCall = fetchMock.mock.calls.find(
      (c) => (c[1] as RequestInit | undefined)?.method === 'POST',
    );
    expect(postCall).toBeTruthy();
    expect(JSON.parse((postCall![1] as RequestInit).body as string)).toMatchObject({
      title: 'New task',
      project_id: 'demo',
    });
  });

  it('edits a work item via the card and PUTs the change', async () => {
    const item = { id: '1', title: 'Old', project_id: 'demo', description_markdown: '', priority: 0 };
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(envelope([item])) // initial list
      .mockResolvedValueOnce(okData({ ...item, title: 'Renamed' })) // PUT
      .mockResolvedValueOnce(envelope([{ ...item, title: 'Renamed' }])); // refetch
    vi.stubGlobal('fetch', fetchMock);

    render(() => <WorkItems />);

    fireEvent.click(await screen.findByRole('button', { name: 'Edit' }));
    fireEvent.input(screen.getByLabelText('Title'), { target: { value: 'Renamed' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    expect(await screen.findByText('Renamed')).toBeInTheDocument();

    const putCall = fetchMock.mock.calls.find(
      (c) => (c[1] as RequestInit | undefined)?.method === 'PUT',
    );
    expect(putCall).toBeTruthy();
    expect(JSON.parse((putCall![1] as RequestInit).body as string)).toMatchObject({ title: 'Renamed' });
  });

  it('shows an error when the API fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        json: async () => ({ status: false, data: null, error: 'boom' }),
      }),
    );

    render(() => <WorkItems />);

    expect(await screen.findByText(/boom/)).toBeInTheDocument();
  });
});
