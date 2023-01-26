import React from 'react';
import styles from './Hero.module.scss'
import Link from '@docusaurus/Link'
import image from '@site/static/img/overview.drawio.png'

const Hero = () => {
    return (
        <div className={styles.main}>
            <div className={styles.mainContent}>

                <div className={styles.description}>Ease of use as the main concern.</div>

                <div className={styles.libName}>Freyja</div>

                <div className={styles.tagLine}>Create and manage networks and virtual machines easily.</div>

                <div className={styles.buttons}>
                    <Link to={'docs/introduction'}>
                        <button className={styles.button}>View Docs</button>
                    </Link>
                </div>


            </div>

            <img src={image} height={"350vh"} alt="showcase"/>

        </div>
    );
};

export default Hero;
