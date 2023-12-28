import React, { useState } from 'react';
import { Button, Input, Spacer } from '@nextui-org/react';
import textContent from "@/i18n";
import iCloudLogo from '@/assets/icloud.jpg'
import { history } from 'umi';
import { addAccount } from "@/net";

const IndexPage = () => {
  const [account, setAccount] = useState('');
  const [password, setPassword] = useState('');
  const [twoFactorCode, setTwoFactorCode] = useState('');
  const [needsTwoFactor, setNeedsTwoFactor] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [language, setLanguage] = useState('cn');

  const texts = language === 'cn' ? textContent.cn : textContent.en;

  const handleSubmit = async () => {
    setIsLoading(true);
    try {
      const {needsTwoFactor, success} = await addAccount(account, password, twoFactorCode)
      if (needsTwoFactor) {
        setNeedsTwoFactor(true);
      } else if (success) {
        // 处理账号添加成功
        history.push('/accountManage')
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
        <h1 style={{fontSize: '24px'}}>{texts.addAccount}</h1>
        <img src={iCloudLogo} width={200} style={{}}/>
        <div style={{padding: 20, maxWidth: '600px', margin: 'auto'}}>
          <Input
            isClearable
            label={texts.accountPlaceholder}
            value={account}
            onChange={(e) => setAccount(e.target.value)}/>
          <Spacer y={1}/>
          <Input
            isClearable
            label={texts.passwordPlaceholder}
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}/>
          {needsTwoFactor && (
            <>
              <Spacer y={1}/>
              <Input
                isClearable
                label={texts.twoFACodePlaceholder}
                value={twoFactorCode}
                onChange={(e) => setTwoFactorCode(e.target.value)}/>
            </>
          )}
          <Spacer y={1.5}/>
          <Button disabled={isLoading} onClick={handleSubmit} color="primary">
            {isLoading ? texts.addingButton : texts.addButton}
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
