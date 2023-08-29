/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import { module, test } from 'qunit';
import { setupRenderingTest } from 'vault/tests/helpers';
import { setupEngine } from 'ember-engines/test-support';
import { setupMirage } from 'ember-cli-mirage/test-support';
import { render, focus, triggerKeyEvent, fillIn } from '@ember/test-helpers';
import { hbs } from 'ember-cli-htmlbars';
import { kvMetadataPath } from 'vault/utils/kv-path';

const MODELS = {
  secrets: [
    {
      id: kvMetadataPath('my-engine', 'my-secret'),
      path: 'my-secret',
      fullSecretPath: 'my-secret',
    },
    {
      id: kvMetadataPath('my-engine', 'my'),
      path: 'my',
      fullSecretPath: 'my',
    },
    {
      id: kvMetadataPath('my-engine', 'beep/boop/bop'),
      path: 'beep/boop/bop',
      fullSecretPath: 'beep/boop/bop',
    },
    {
      id: kvMetadataPath('my-engine', 'beep/boop-1'),
      path: 'beep/boop-1',
      fullSecretPath: 'beep/boop-1',
    },
  ],
};
const MOUNT_POINT = 'vault.cluster.secrets.backend.kv';

module('Integration | Component | kv | kv-list-filter', function (hooks) {
  setupRenderingTest(hooks);
  setupEngine(hooks, 'kv');
  setupMirage(hooks);

  hooks.beforeEach(function () {
    this.model = MODELS;
    this.mountPoint = MOUNT_POINT;
  });

  test('it clears last item on backspace and clears to directory on esc', async function (assert) {
    assert.expect(8);
    // mirage hook for filling in the input
    this.owner.lookup('service:router').reopen({
      transitionTo(route, pathToSecret, { queryParams: { pageFilter } }) {
        assert.deepEqual(pageFilter, 'boop-', 'Sends the correct pageFilter on fillIn.');
      },
    });

    await render(
      hbs`<KvListFilter @secrets={{this.model.secrets}} @mountPoint={{this.mountPoint}} @filterValue="beep/" @pageFilter=""/>`,
      {
        owner: this.engine,
      }
    );
    // focus on input and trigger backspace
    await focus('[data-test-component="kv-list-filter"]');
    await fillIn('[data-test-component="kv-list-filter"]', 'beep/boop-');

    this.owner.lookup('service:router').reopen({
      transitionTo(route, pathToSecret, { queryParams: { pageFilter } }) {
        assert.strictEqual(route, `${MOUNT_POINT}.list-directory`, 'Correct route sent.');
        assert.strictEqual(pathToSecret, 'beep/', 'PathToSecret is the parent directory.');
        assert.deepEqual(pageFilter, 'boop', 'Clears last item in pageFilter on backspace.');
      },
    });
    await triggerKeyEvent('[data-test-component="kv-list-filter"]', 'keydown', 8);
    assert.strictEqual(
      document.activeElement.id,
      'secret-filter',
      'the input still remains focused after delete.'
    );

    this.owner.lookup('service:router').reopen({
      transitionTo(route, pathToSecret, { queryParams: { pageFilter } }) {
        assert.strictEqual(route, `${MOUNT_POINT}.list-directory`, 'Still on a directory route.');
        assert.strictEqual(pathToSecret, 'beep/', 'Parent directory still shown.');
        assert.deepEqual(pageFilter, null, 'Clears pageFilter on escape.');
      },
    });
    // trigger escape
    await triggerKeyEvent('[data-test-component="kv-list-filter"]', 'keydown', 27);
  });
});
