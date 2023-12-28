import React, { useState } from 'react';
import { Button, Input, Spacer } from '@nextui-org/react';
import textContent from "@/i18n";
import { history } from 'umi';
import { login } from "@/net";

const IndexPage = () => {
  const [account, setAccount] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [language, setLanguage] = useState('cn');

  const texts = language === 'cn' ? textContent.cn : textContent.en;

  const handleSubmit = async () => {
    setIsLoading(true);
    try {
      const {success} = await login(account, password)
      if (success) {
        // 处理账号添加成功
        history.push('/accountManager')
      }
    } catch (error) {
      // 处理错误
    }
    setIsLoading(false);
  };

  const toggleLanguage = () => {
    setLanguage(language === 'cn' ? 'en' : 'cn');
  };


  return (
    <div style={{textAlign: 'center'}}>
      <div style={{padding: 20, maxWidth: '600px', margin: 'auto'}}>
        <h1 style={{fontSize: '24px'}}>{texts.loginTitle}</h1>
        <div style={{padding: 20, maxWidth: '600px', margin: 'auto'}}>
          <Input
            isClearable
            label={texts.loginAccountPlaceholder}
            value={account}
            onChange={(e) => setAccount(e.target.value)}/>
          <Spacer y={1}/>
          <Input
            isClearable
            label={texts.loginPasswordPlaceholder}
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}/>
          <Spacer y={1.5}/>
          <Button disabled={isLoading} onClick={handleSubmit} color="primary">
            {isLoading ? texts.loginingButton : texts.loginButton}
          </Button>
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

export default IndexPage;
