import React, { useState } from 'react';
import { Button, Input, Spacer } from '@nextui-org/react';
import axios from 'axios';
import textContent from "@/i18n";
import { useLocation } from 'umi';

export default () => {
  const location = useLocation();
  const emailAccount = location.pathname.split('configure/')[1]
  console.log('emailAccount', emailAccount)
  const [account, setAccount] = useState('x@xx.com');
  const [password, setPassword] = useState('');
  const [folderFormat, setFolderFormat] = useState('2006/01/02');
  const [removeDeleted, setRemoveDeleted] = useState('是'); // 修改为字符串，默认为“否”
  const [concurrency, setConcurrency] = useState(10);
  const [language, setLanguage] = useState('cn');

  const texts = language === 'cn' ? textContent.cn : textContent.en;

  const handleSubmit = async () => {
    // 这里发送配置信息的 HTTP 请求
    try {
      await axios.post('/api/config', {account, password, folderFormat, removeDeleted, concurrency});
      // 处理成功响应
    } catch (error) {
      // 处理错误
    }
  };

  const handleCancel = () => {
    // 返回上一页
    window.history.back();
  };

  const toggleLanguage = () => {
    setLanguage(language === 'cn' ? 'en' : 'cn');
  };

  return (
    <div style={{textAlign: 'center'}}>
      <div style={{padding: 20, maxWidth: '600px', margin: 'auto'}}>
        <h1 style={{fontSize: '24px', marginBottom: '20px'}}>{texts.icloudPhotoDownloadConfig}</h1>
        <Input
          // isClearable
          isReadOnly
          label={texts.iCloudAccount}
          value={emailAccount}
          onChange={(e) => setAccount(e.target.value)}
        />
        <Spacer y={1}/>
        <Input
          isClearable
          label={texts.iCloudPassword}
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <Spacer y={1}/>
        <Input
          isClearable
          label={texts.folderStruct}
          value={folderFormat}
          onChange={(e) => setFolderFormat(e.target.value)}
        />
        <Spacer y={1}/>
        <Input
          isClearable
          label={texts.threadNum}
          type="number"
          value={`${concurrency}`}
          onChange={(e) => setConcurrency(parseInt(e.target.value, 10))}
        />
        <Spacer y={1.5}/>
        <div style={{display: 'flex', justifyContent: 'center', gap: '20px'}}>
          <Button color="primary" onClick={handleSubmit}>{texts.save}</Button>
          <Button color="default" onClick={handleCancel}>{texts.cancel}</Button>
        </div>
      </div>
      <div style={{display: 'flex', justifyContent: 'center', marginTop: '20px'}}>
        <span onClick={toggleLanguage} style={{cursor: 'pointer'}}>
          {language === 'cn' ? '🇨🇳' : '🇺🇸'} {texts.toggleLanguage}
        </span>
      </div>
    </div>
  );
};

