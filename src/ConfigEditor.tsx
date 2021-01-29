import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions } from './types';

const { FormField } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onGuildIDChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      discordGuildID: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onGuildIDKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      discordGuildIDKey: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onResetAPIKey = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        apiKey: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        apiKey: '',
      },
    });
  };

  render() {
    const { options } = this.props;
    const { jsonData } = options;
    //const secureJsonData = (options.secureJsonData || {}) as MySecureJsonData;

    return (
      <div className="gf-form-group">
        <div className="gf-form">
          <FormField
            label="DiscordGuildID"
            labelWidth={10}
            inputWidth={30}
            onChange={this.onGuildIDChange}
            value={jsonData.discordGuildID || ''}
            placeholder="Discord Guild ID"
          />

          <FormField
            label="DiscordGuildIDKey"
            labelWidth={10}
            inputWidth={30}
            onChange={this.onGuildIDKeyChange}
            value={jsonData.discordGuildIDKey || ''}
            placeholder="Discord Guild ID Key"
          />
        </div>
      </div>
    );
  }
}
