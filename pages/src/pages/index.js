import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';

import styles from './index.module.scss';
import Hero from "../components/Hero/Hero";
import Usage from "../components/Usage/Usage";
import Tech from "../components/Tech/Tech";

export default function Home() {
    const {siteConfig} = useDocusaurusContext();
    return (
        <Layout
            wrapperClassName={styles.wrapper}
            title={`Hello from ${siteConfig.title}`}
            description="Description will go into a meta tag in <head />">
            <main>
                <Hero/>
                <Usage/>
                <Tech/>
            </main>

        </Layout>
    );
}
